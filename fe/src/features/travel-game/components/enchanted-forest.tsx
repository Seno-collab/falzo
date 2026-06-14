"use client";

import { useFrame } from "@react-three/fiber";
import { useMemo, useRef } from "react";
import type { Group } from "three";

type EnchantedForestProps = {
  selected: boolean;
  variant: number;
};

const treeLayouts = [
  [
    [-0.42, 0.12, -0.2, 0.82],
    [-0.3, 0.13, 0.1, 0.64],
    [-0.1, 0.12, -0.34, 0.58],
    [0.14, 0.12, 0.24, 0.52],
    [0.36, 0.12, -0.12, 0.7],
    [0.43, 0.12, 0.18, 0.48],
    [-0.02, 0.12, 0.0, 0.62],
  ],
  [
    [-0.38, 0.12, 0.18, 0.74],
    [-0.2, 0.12, -0.25, 0.56],
    [0.02, 0.12, 0.32, 0.62],
    [0.2, 0.12, -0.1, 0.5],
    [0.38, 0.12, -0.24, 0.66],
    [0.42, 0.12, 0.14, 0.48],
    [-0.04, 0.12, -0.02, 0.68],
  ],
  [
    [-0.46, 0.12, 0.02, 0.68],
    [-0.28, 0.12, -0.24, 0.58],
    [-0.16, 0.12, 0.28, 0.52],
    [0.08, 0.12, -0.34, 0.64],
    [0.24, 0.12, 0.22, 0.58],
    [0.42, 0.12, -0.02, 0.72],
    [0.04, 0.12, 0.02, 0.54],
  ],
];

export function EnchantedForest({
  selected,
  variant,
}: Readonly<EnchantedForestProps>) {
  const groupRef = useRef<Group>(null);
  const trees = useMemo(
    () => treeLayouts[variant % treeLayouts.length],
    [variant],
  );

  useFrame((state) => {
    if (!groupRef.current) {
      return;
    }

    groupRef.current.rotation.y =
      Math.sin(state.clock.elapsedTime * 0.5 + variant) * 0.035;
  });

  return (
    <group ref={groupRef}>
      {trees.map(([x, y, z, scale], index) => (
        <group key={`${x}-${z}`} position={[x, y, z]} scale={scale}>
          <mesh position={[0, 0.08, 0]}>
            <cylinderGeometry args={[0.035, 0.048, 0.18, 8]} />
            <meshStandardMaterial
              color="#67482d"
              emissive="#2a1710"
              emissiveIntensity={0.08}
              roughness={0.78}
            />
          </mesh>
          <mesh position={[0, 0.23, 0]}>
            <coneGeometry args={[0.15, 0.34, 10]} />
            <meshStandardMaterial
              color={selected ? "#2f9a64" : "#236d4f"}
              emissive={selected ? "#124a31" : "#09291f"}
              emissiveIntensity={selected ? 0.2 : 0.09}
              roughness={0.68}
            />
          </mesh>
          <mesh position={[0, 0.4, 0]}>
            <coneGeometry args={[0.11, 0.27, 10]} />
            <meshStandardMaterial
              color={selected ? "#5bb980" : "#2f805c"}
              emissive={selected ? "#1d5638" : "#0d3325"}
              emissiveIntensity={selected ? 0.16 : 0.08}
              roughness={0.62}
            />
          </mesh>
          <mesh position={[index % 2 === 0 ? 0.12 : -0.11, 0.02, 0.08]}>
            <sphereGeometry args={[0.045, 12, 8]} />
            <meshStandardMaterial
              color={selected ? "#6cb36d" : "#3f7a4c"}
              roughness={0.7}
            />
          </mesh>
        </group>
      ))}
    </group>
  );
}
