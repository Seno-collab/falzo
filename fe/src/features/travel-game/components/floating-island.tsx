"use client";

type FloatingIslandProps = {
  selected: boolean;
};

export function FloatingIsland({ selected }: Readonly<FloatingIslandProps>) {
  return (
    <group scale={selected ? 1.08 : 1}>
      <mesh position={[0, -0.28, 0]} rotation={[Math.PI, 0, 0]}>
        <coneGeometry args={[0.7, 0.72, 11]} />
        <meshStandardMaterial
          color={selected ? "#4f432c" : "#3d3326"}
          emissive={selected ? "#2a2417" : "#17130d"}
          emissiveIntensity={0.14}
          roughness={0.86}
        />
      </mesh>
      <mesh position={[0, -0.03, 0]} scale={[1.22, 0.34, 0.86]}>
        <sphereGeometry args={[0.58, 28, 14]} />
        <meshStandardMaterial
          color={selected ? "#2f7c56" : "#285f4a"}
          emissive={selected ? "#123d2d" : "#09251e"}
          emissiveIntensity={selected ? 0.2 : 0.1}
          roughness={0.78}
        />
      </mesh>
      <mesh position={[0, 0.04, 0.08]} scale={[1.05, 0.12, 0.62]}>
        <sphereGeometry args={[0.46, 24, 10]} />
        <meshStandardMaterial
          color={selected ? "#d7bd7a" : "#aa9562"}
          emissive="#2b2315"
          emissiveIntensity={0.06}
          roughness={0.82}
        />
      </mesh>
      <mesh position={[-0.22, 0.11, -0.04]} scale={[0.48, 0.12, 0.34]}>
        <sphereGeometry args={[0.42, 18, 8]} />
        <meshStandardMaterial
          color={selected ? "#3d9867" : "#2f724e"}
          emissive={selected ? "#16432f" : "#0b2d21"}
          emissiveIntensity={selected ? 0.18 : 0.08}
          roughness={0.76}
        />
      </mesh>
      {selected ? (
        <mesh position={[0, 0.11, 0]} rotation={[-Math.PI / 2, 0, 0]}>
          <ringGeometry args={[0.68, 0.72, 64]} />
          <meshBasicMaterial color="#f7c948" transparent opacity={0.28} />
        </mesh>
      ) : null}
    </group>
  );
}
