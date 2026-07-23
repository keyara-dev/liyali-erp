"use client";
import { useCallback } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

export interface UseDateRangeUrlStateOptions {
  /** Fallback ISO `YYYY-MM-DD` when `?from` is absent. */
  defaultFrom: string;
  /** Fallback ISO `YYYY-MM-DD` when `?to` is absent. */
  defaultTo: string;
  /** Param key for from (default "from"). */
  fromKey?: string;
  /** Param key for to (default "to"). */
  toKey?: string;
}

export interface UseDateRangeUrlStateResult {
  from: string;
  to: string;
  setRange: (from: string, to: string) => void;
}

export function useDateRangeUrlState({
  defaultFrom,
  defaultTo,
  fromKey = "from",
  toKey = "to",
}: UseDateRangeUrlStateOptions): UseDateRangeUrlStateResult {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const searchString = searchParams.toString();

  const from = searchParams.get(fromKey) || defaultFrom;
  const to = searchParams.get(toKey) || defaultTo;

  const setRange = useCallback(
    (newFrom: string, newTo: string) => {
      const params = new URLSearchParams(searchString);
      const currentFrom = params.get(fromKey) || defaultFrom;
      const currentTo = params.get(toKey) || defaultTo;
      if (newFrom === currentFrom && newTo === currentTo) {
        return;
      }
      params.set(fromKey, newFrom);
      params.set(toKey, newTo);
      router.replace(`${pathname}?${params.toString()}`, { scroll: false });
    },
    [router, pathname, searchString, fromKey, toKey, defaultFrom, defaultTo]
  );

  return { from, to, setRange };
}
