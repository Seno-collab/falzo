"use client";

import { useFrame } from "@react-three/fiber";
import { useRef } from "react";
import type { Group } from "three";

export function MissionPortal({ active }: Readonly<{ active: boolean }>) {
  const portalRef = useRef<Group>(null);

  useFrame((_, delta) => {
    if (!portalRef.current) {
      return;
    }

    portalRef.current.rotation.y += delta * (active ? 0.9 : 0.32);
    portalRef.current.rotation.z += delta * (active ? 0.22 : 0.08);
  });

  return (
    <group ref={portalRef} position={[0, 0.42, 0]}>
      <mesh rotation={[Math.PI / 2, 0, 0]}>
        <torusGeometry args={[0.42, 0.018, 12, 72]} />
        <meshBasicMaterial
          color={active ? "#f7c948" : "#45d89d"}
          transparent
          opacity={active ? 0.95 : 0.42}
        />
      </mesh>
      <mesh rotation={[Math.PI / 2, Math.PI / 3, 0]}>
        <torusGeometry args={[0.3, 0.014, 12, 72]} />
        <meshBasicMaterial
          color={active ? "#fff1b8" : "#9fe7cb"}
          transparent
          opacity={active ? 0.78 : 0.32}
        />
      </mesh>
      {active ? (
        <pointLight color="#f7c948" intensity={2.8} distance={2.4} />
      ) : null}
    </group>
  );
}
