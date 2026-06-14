"use client";

import { useFrame } from "@react-three/fiber";
import { useRef } from "react";
import type { Group } from "three";

const orbPositions = [
  [0.58, 0.7, 0] as const,
  [-0.3, 0.86, 0.5] as const,
  [-0.24, 0.62, -0.58] as const,
];

export function XpOrb() {
  const groupRef = useRef<Group>(null);

  useFrame((state, delta) => {
    if (!groupRef.current) {
      return;
    }

    groupRef.current.rotation.y += delta * 1.35;
    groupRef.current.position.y = Math.sin(state.clock.elapsedTime * 1.8) * 0.05;
  });

  return (
    <group ref={groupRef}>
      {orbPositions.map((position, index) => (
        <mesh key={`${position.join(":")}-${index}`} position={position}>
          <sphereGeometry args={[index === 0 ? 0.08 : 0.06, 18, 18]} />
          <meshStandardMaterial
            color="#f7c948"
            emissive="#f7c948"
            emissiveIntensity={1.35}
            roughness={0.28}
          />
        </mesh>
      ))}
    </group>
  );
}
