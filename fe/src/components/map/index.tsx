import dynamic from "next/dynamic";
import styles from "./map.module.css";
import type { FalzoMapProps } from "./types";

const MapClient = dynamic<FalzoMapProps>(() => import("./map"), {
  loading: () => <div className={styles.loading}>Loading map</div>,
  ssr: false,
});

export default MapClient;
export type { Coordinates, FalzoMapProps, MapPoint } from "./types";
