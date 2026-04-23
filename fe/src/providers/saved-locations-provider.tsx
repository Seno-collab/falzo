"use client";

import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from "react";
import type { TravelLocation } from "@/lib/travel/types";

type SavedLocationsContextValue = {
  savedLocations: TravelLocation[];
  isSaved: (id: string) => boolean;
  toggleSaved: (location: TravelLocation) => void;
  removeSaved: (id: string) => void;
};

const STORAGE_KEY = "travel.saved_locations";
const SavedLocationsContext = createContext<SavedLocationsContextValue | null>(null);

function normalizeSaved(payload: unknown): TravelLocation[] {
  if (!Array.isArray(payload)) {
    return [];
  }

  return payload.filter((item): item is TravelLocation => {
    return Boolean(
      item &&
        typeof item === "object" &&
        typeof (item as TravelLocation).id === "string" &&
        typeof (item as TravelLocation).name === "string",
    );
  });
}

export function SavedLocationsProvider({ children }: PropsWithChildren) {
  const [savedLocations, setSavedLocations] = useState<TravelLocation[]>([]);

  useEffect(() => {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return;
    }

    try {
      setSavedLocations(normalizeSaved(JSON.parse(raw)));
    } catch {
      setSavedLocations([]);
    }
  }, []);

  useEffect(() => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(savedLocations));
  }, [savedLocations]);

  const value = useMemo<SavedLocationsContextValue>(() => {
    return {
      savedLocations,
      isSaved: (id) => savedLocations.some((item) => item.id === id),
      removeSaved: (id) => {
        setSavedLocations((previous) => previous.filter((item) => item.id !== id));
      },
      toggleSaved: (location) => {
        setSavedLocations((previous) => {
          const exists = previous.some((item) => item.id === location.id);
          if (exists) {
            return previous.filter((item) => item.id !== location.id);
          }

          return [location, ...previous];
        });
      },
    };
  }, [savedLocations]);

  return (
    <SavedLocationsContext.Provider value={value}>{children}</SavedLocationsContext.Provider>
  );
}

export function useSavedLocations() {
  const context = useContext(SavedLocationsContext);

  if (!context) {
    throw new Error("useSavedLocations must be used within SavedLocationsProvider");
  }

  return context;
}
