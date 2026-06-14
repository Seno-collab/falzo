"use client";

import { Float, Html, Line, OrbitControls, Sparkles } from "@react-three/drei";
import { Canvas, useFrame } from "@react-three/fiber";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { Group } from "three";
import { Color, Vector3 } from "three";
import { EnchantedForest } from "@/features/travel-game/components/enchanted-forest";
import { FloatingIsland } from "@/features/travel-game/components/floating-island";
import { TravelerFigure } from "@/features/travel-game/components/traveler-figure";
import type {
  DestinationNodeId,
  TravelGameCopy,
} from "@/features/travel-game/data/travel-game-copy";

type TravelGameNode = {
  id: DestinationNodeId;
  imageId: string;
  position: readonly [number, number, number];
};

type SceneProps = {
  copy: TravelGameCopy;
  nodes: readonly TravelGameNode[];
  selectedNodeId: DestinationNodeId;
  onSelectNode: (nodeId: DestinationNodeId) => void;
};

type MoveDirection = "up" | "down" | "left" | "right";

type ActiveDirections = ReadonlySet<MoveDirection>;

const routeColor = new Color("#9fe7cb");
const routeGlowColor = new Color("#f7c948");
const movementBounds = {
  minX: -3.18,
  maxX: 3.18,
  minZ: -0.88,
  maxZ: 0.88,
};
const destinationSnapDistance = 0.48;
const keyboardDirections: Record<string, MoveDirection> = {
  ArrowUp: "up",
  KeyW: "up",
  ArrowDown: "down",
  KeyS: "down",
  ArrowLeft: "left",
  KeyA: "left",
  ArrowRight: "right",
  KeyD: "right",
};

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value));
}

function isTextInputTarget(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) {
    return false;
  }

  const tagName = target.tagName.toLowerCase();
  return (
    tagName === "input" ||
    tagName === "textarea" ||
    tagName === "select" ||
    target.isContentEditable
  );
}

function TravelMap({
  activeDirections,
  copy,
  nodes,
  selectedNodeId,
  onSelectNode,
}: Readonly<SceneProps & { activeDirections: ActiveDirections }>) {
  const groupRef = useRef<Group>(null);
  const selectedNode =
    nodes.find((node) => node.id === selectedNodeId) ?? nodes[0];
  const [playerPosition, setPlayerPosition] = useState<
    readonly [number, number, number]
  >(() => selectedNode.position);
  const [playerHeading, setPlayerHeading] = useState(-0.45);
  const [playerMoving, setPlayerMoving] = useState(false);
  const routePoints = useMemo(
    () => nodes.map((node) => new Vector3(...node.position)),
    [nodes],
  );
  const nodeVectors = useMemo(
    () =>
      nodes.map((node) => ({
        id: node.id,
        position: new Vector3(...node.position),
      })),
    [nodes],
  );

  useEffect(() => {
    setPlayerPosition(selectedNode.position);
  }, [selectedNode.position]);

  useFrame((state, delta) => {
    if (!groupRef.current) {
      return;
    }

    groupRef.current.rotation.y =
      Math.sin(state.clock.elapsedTime * 0.25) * 0.08;
    groupRef.current.rotation.x =
      -0.08 + Math.sin(state.clock.elapsedTime * 0.18) * 0.025;

    const xAxis =
      (activeDirections.has("right") ? 1 : 0) -
      (activeDirections.has("left") ? 1 : 0);
    const zAxis =
      (activeDirections.has("down") ? 1 : 0) -
      (activeDirections.has("up") ? 1 : 0);
    const moving = xAxis !== 0 || zAxis !== 0;
    setPlayerMoving(moving);

    if (!moving) {
      return;
    }

    setPlayerPosition((currentPosition) => {
      const moveVector = new Vector3(xAxis, 0, zAxis).normalize();
      const nextX = clamp(
        currentPosition[0] + moveVector.x * delta * 1.45,
        movementBounds.minX,
        movementBounds.maxX,
      );
      const nextZ = clamp(
        currentPosition[2] + moveVector.z * delta * 1.15,
        movementBounds.minZ,
        movementBounds.maxZ,
      );
      let nearestNode = nodeVectors[0];
      let nearestDistance = Number.POSITIVE_INFINITY;

      for (const node of nodeVectors) {
        const distance = Math.hypot(
          node.position.x - nextX,
          node.position.z - nextZ,
        );
        if (distance < nearestDistance) {
          nearestDistance = distance;
          nearestNode = node;
        }
      }

      if (
        nearestDistance < destinationSnapDistance &&
        nearestNode.id !== selectedNodeId
      ) {
        onSelectNode(nearestNode.id);
      }

      setPlayerHeading(Math.atan2(moveVector.x, moveVector.z) + Math.PI);

      return [
        nextX,
        currentPosition[1] + (nearestNode.position.y - currentPosition[1]) * 0.08,
        nextZ,
      ] as const;
    });
  });

  return (
    <group ref={groupRef} position={[0.7, 0.02, 0]}>
      <Line
        color={routeGlowColor}
        dashed={false}
        lineWidth={8}
        points={routePoints}
        transparent
        opacity={0.16}
      />
      <Line
        color={routeColor}
        dashed={false}
        lineWidth={3.2}
        points={routePoints}
        transparent
        opacity={0.9}
      />

      <mesh position={[0, -0.72, 0]} rotation={[-Math.PI / 2, 0, 0]}>
        <circleGeometry args={[4.2, 96]} />
        <meshStandardMaterial
          color="#0e2525"
          emissive="#102b33"
          emissiveIntensity={0.94}
          metalness={0.08}
          roughness={0.74}
          transparent
          opacity={0.7}
        />
      </mesh>

      <mesh position={[0, -0.76, 0]} rotation={[-Math.PI / 2, 0, 0]}>
        <ringGeometry args={[4.25, 4.32, 128]} />
        <meshStandardMaterial
          color="#f7c948"
          emissive="#8d6a0f"
          emissiveIntensity={0.48}
          transparent
          opacity={0.62}
        />
      </mesh>

      {nodes.map((node, index) => {
        const destination = copy.destinations[node.id];
        const selected = node.id === selectedNodeId;
        const color = selected ? "#f7c948" : "#45d89d";
        const scale = selected ? 1 : 0.82;

        return (
          <Float
            floatIntensity={selected ? 0.16 : 0.08}
            key={node.id}
            rotationIntensity={selected ? 0.08 : 0.04}
            speed={0.55 + index * 0.04}
          >
            <group position={node.position}>
              <FloatingIsland selected={selected} />
              <EnchantedForest selected={selected} variant={index} />
              <mesh
                aria-label={`${copy.nodeLabel}: ${destination.name}`}
                onClick={(event) => {
                  event.stopPropagation();
                  onSelectNode(node.id);
                }}
                onPointerEnter={(event) => {
                  event.stopPropagation();
                  document.body.style.cursor = "pointer";
                }}
                onPointerLeave={() => {
                  document.body.style.cursor = "";
                }}
                position={[0, 0.2, 0]}
                scale={scale}
              >
                <sphereGeometry args={[0.12, 18, 12]} />
                <meshStandardMaterial
                  color={color}
                  emissive={color}
                  emissiveIntensity={selected ? 0.54 : 0.18}
                  metalness={0.04}
                  roughness={0.58}
                  transparent
                  opacity={selected ? 0.92 : 0.64}
                />
              </mesh>
              {selected ? (
                <mesh position={[0, 0.12, 0]} rotation={[-Math.PI / 2, 0, 0]}>
                  <ringGeometry args={[0.48, 0.5, 48]} />
                  <meshBasicMaterial color={color} transparent opacity={0.34} />
                </mesh>
              ) : null}
              <Html
                center
                className="pointer-events-none select-none"
                distanceFactor={selected ? 5.8 : 6.4}
                position={[0, 1.05, 0]}
              >
                <div
                  className={[
                    "whitespace-nowrap rounded-lg border px-3 py-1.5 text-center shadow-[0_12px_28px_-18px_rgb(0_0_0/0.9)] backdrop-blur",
                    selected
                      ? "border-[#f7c948]/60 bg-[#f7c948]/95 text-[#16312a]"
                      : "border-white/20 bg-[#09211c]/72 text-white",
                  ].join(" ")}
                >
                  <span className="block text-[0.62rem] font-black uppercase tracking-[0.08em]">
                    {destination.name}
                  </span>
                  {selected ? (
                    <span className="block text-[0.68rem] font-bold">
                      {destination.realm}
                    </span>
                  ) : null}
                </div>
              </Html>
            </group>
          </Float>
        );
      })}
      <TravelerFigure
        heading={playerHeading}
        moving={playerMoving}
        position={[
          playerPosition[0] - 0.2,
          playerPosition[1] + 0.36,
          playerPosition[2] + 0.08,
        ]}
      />
    </group>
  );
}

function ControlButton({
  direction,
  label,
  onDirectionChange,
}: Readonly<{
  direction: MoveDirection;
  label: string;
  onDirectionChange: (direction: MoveDirection, active: boolean) => void;
}>) {
  return (
    <button
      aria-label={label}
      className="inline-flex size-9 touch-none select-none items-center justify-center rounded-lg border border-white/18 bg-white/16 text-sm font-black text-white shadow-[0_14px_32px_-22px_rgb(0_0_0/0.9)] backdrop-blur transition hover:bg-white/24 active:scale-95"
      onPointerCancel={() => onDirectionChange(direction, false)}
      onPointerDown={(event) => {
        event.currentTarget.setPointerCapture(event.pointerId);
        onDirectionChange(direction, true);
      }}
      onPointerLeave={() => onDirectionChange(direction, false)}
      onPointerUp={(event) => {
        event.currentTarget.releasePointerCapture(event.pointerId);
        onDirectionChange(direction, false);
      }}
      type="button"
    >
      {label}
    </button>
  );
}

export function TravelGameScene(props: Readonly<SceneProps>) {
  const [activeDirections, setActiveDirections] = useState<
    Set<MoveDirection>
  >(() => new Set());
  const setDirectionActive = useCallback(
    (direction: MoveDirection, active: boolean) => {
      setActiveDirections((currentDirections) => {
        const nextDirections = new Set(currentDirections);
        if (active) {
          nextDirections.add(direction);
        } else {
          nextDirections.delete(direction);
        }
        return nextDirections;
      });
    },
    [],
  );

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (isTextInputTarget(event.target)) {
        return;
      }

      const direction = keyboardDirections[event.code];
      if (!direction) {
        return;
      }

      event.preventDefault();
      setDirectionActive(direction, true);
    }

    function handleKeyUp(event: KeyboardEvent) {
      const direction = keyboardDirections[event.code];
      if (!direction) {
        return;
      }

      event.preventDefault();
      setDirectionActive(direction, false);
    }

    globalThis.addEventListener("keydown", handleKeyDown);
    globalThis.addEventListener("keyup", handleKeyUp);

    return () => {
      globalThis.removeEventListener("keydown", handleKeyDown);
      globalThis.removeEventListener("keyup", handleKeyUp);
    };
  }, [setDirectionActive]);

  return (
    <div
      aria-label={props.copy.sceneLabel}
      className="absolute inset-0"
      role="img"
    >
      <Canvas
        camera={{ position: [0, 2.65, 5.65], fov: 48 }}
        dpr={[1, 1.65]}
        gl={{ alpha: true, antialias: true, powerPreference: "high-performance" }}
      >
        <fog args={["#061617", 5.5, 11]} attach="fog" />
        <ambientLight intensity={0.68} />
        <directionalLight
          color="#fff1b8"
          intensity={2.35}
          position={[-3, 4, 5]}
        />
        <directionalLight color="#8fd8b6" intensity={0.85} position={[5, 2, -4]} />
        <pointLight color="#9fe7cb" intensity={5.2} position={[2.8, 1.2, 2.5]} />
        <pointLight color="#f7c948" intensity={3.4} position={[-2.6, 1.1, 2]} />
        <Sparkles
          color="#f7d36b"
          count={28}
          opacity={0.24}
          scale={[8.6, 4.2, 5.4]}
          size={2.2}
          speed={0.12}
        />
        <TravelMap {...props} activeDirections={activeDirections} />
        <OrbitControls
          autoRotate
          autoRotateSpeed={0.18}
          enablePan={false}
          enableZoom={false}
          maxPolarAngle={Math.PI / 2.05}
          minPolarAngle={Math.PI / 3.2}
        />
      </Canvas>
      <div className="pointer-events-none absolute bottom-4 right-4 hidden max-w-48 rounded-lg border border-white/12 bg-[#061817]/62 p-3 text-white shadow-[0_18px_42px_-28px_rgb(0_0_0/0.9)] backdrop-blur-xl lg:block">
        <p className="mb-2 text-[0.65rem] font-black uppercase tracking-[0.12em] text-[#9fe7cb]">
          {props.copy.controlsLabel}
        </p>
        <p className="text-xs font-semibold leading-4 text-white/68">
          {props.copy.controlsHint}
        </p>
      </div>
      <div className="absolute bottom-4 left-4 grid grid-cols-3 gap-1.5 lg:hidden">
        <span />
        <ControlButton
          direction="up"
          label={props.copy.moveUpLabel}
          onDirectionChange={setDirectionActive}
        />
        <span />
        <ControlButton
          direction="left"
          label={props.copy.moveLeftLabel}
          onDirectionChange={setDirectionActive}
        />
        <ControlButton
          direction="down"
          label={props.copy.moveDownLabel}
          onDirectionChange={setDirectionActive}
        />
        <ControlButton
          direction="right"
          label={props.copy.moveRightLabel}
          onDirectionChange={setDirectionActive}
        />
      </div>
    </div>
  );
}
