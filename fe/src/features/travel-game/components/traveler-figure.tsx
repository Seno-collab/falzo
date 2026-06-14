"use client";

import { useFrame } from "@react-three/fiber";
import { useRef } from "react";
import type { Group } from "three";

type TravelerFigureProps = {
  heading?: number;
  moving?: boolean;
  position?: readonly [number, number, number];
};

export function TravelerFigure({
  heading = -0.45,
  moving = false,
  position = [-0.2, 0.36, 0.08],
}: Readonly<TravelerFigureProps>) {
  const groupRef = useRef<Group>(null);

  useFrame((state) => {
    if (!groupRef.current) {
      return;
    }

    groupRef.current.position.y =
      position[1] +
      Math.sin(state.clock.elapsedTime * (moving ? 6.4 : 2.1)) *
        (moving ? 0.042 : 0.018);
    groupRef.current.rotation.y =
      heading + Math.sin(state.clock.elapsedTime * 0.9) * 0.08;
  });

  return (
    <group ref={groupRef} position={position} scale={0.56}>
      <mesh position={[0, 0.34, 0]}>
        <sphereGeometry args={[0.13, 24, 18]} />
        <meshStandardMaterial
          color="#ffd2a6"
          emissive="#5a2b18"
          emissiveIntensity={0.08}
          roughness={0.5}
        />
      </mesh>
      <mesh position={[0, 0.5, 0]}>
        <coneGeometry args={[0.17, 0.16, 18]} />
        <meshStandardMaterial
          color="#f7c948"
          emissive="#a87510"
          emissiveIntensity={0.28}
          roughness={0.46}
        />
      </mesh>
      <mesh position={[0, 0.13, 0]}>
        <cylinderGeometry args={[0.12, 0.15, 0.28, 16]} />
        <meshStandardMaterial
          color="#2bd28f"
          emissive="#137653"
          emissiveIntensity={0.22}
          metalness={0.08}
          roughness={0.6}
        />
      </mesh>
      <mesh position={[0, 0.1, -0.12]}>
        <boxGeometry args={[0.17, 0.22, 0.06]} />
        <meshStandardMaterial
          color="#815d38"
          emissive="#2f1c12"
          emissiveIntensity={0.12}
          roughness={0.7}
        />
      </mesh>
      <mesh position={[-0.12, -0.1, 0]} rotation={[0, 0, -0.12]}>
        <cylinderGeometry args={[0.035, 0.04, 0.24, 10]} />
        <meshStandardMaterial color="#173c34" roughness={0.7} />
      </mesh>
      <mesh position={[0.12, -0.1, 0]} rotation={[0, 0, 0.12]}>
        <cylinderGeometry args={[0.035, 0.04, 0.24, 10]} />
        <meshStandardMaterial color="#173c34" roughness={0.7} />
      </mesh>
      <mesh position={[-0.18, 0.13, 0.02]} rotation={[0.1, 0, 0.58]}>
        <cylinderGeometry args={[0.026, 0.03, 0.22, 10]} />
        <meshStandardMaterial color="#ffd2a6" roughness={0.56} />
      </mesh>
      <mesh position={[0.18, 0.13, 0.02]} rotation={[0.1, 0, -0.58]}>
        <cylinderGeometry args={[0.026, 0.03, 0.22, 10]} />
        <meshStandardMaterial color="#ffd2a6" roughness={0.56} />
      </mesh>
      <mesh position={[0.05, 0.36, 0.11]}>
        <sphereGeometry args={[0.018, 12, 8]} />
        <meshBasicMaterial color="#173c34" />
      </mesh>
      <mesh position={[-0.05, 0.36, 0.11]}>
        <sphereGeometry args={[0.018, 12, 8]} />
        <meshBasicMaterial color="#173c34" />
      </mesh>
      <pointLight color="#f7c948" distance={1.8} intensity={0.9} />
    </group>
  );
}
