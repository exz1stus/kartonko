"use client";

import { create, UseBoundStore, StoreApi } from "zustand";
import {
    beginSelection,
    extendSelection,
    SelectionGesture,
    SpatialSelectionBounds,
} from "@/lib/image/rangeSelection";

export interface GallerySelection {
    selectedIds: number[];
    draggingIds: number[];
    selectionMode: boolean;
    selectionGesture: SelectionGesture<number> | null;
    anchorId: number | null;
    selectionBounds: SpatialSelectionBounds | null;
    toggle: (id: number) => void;
    selectClick: (
        order: number[],
        id: number,
        shift: boolean,
        additive: boolean,
    ) => void;
    selectMany: (ids: number[]) => void;
    beginSelection: (
        order: number[],
        id: number,
        shift: boolean,
        additive: boolean,
    ) => void;
    extendSelection: (id: number) => void;
    beginSpatialSelection: (
        ids: number[],
        start: number,
        target: number,
        additive: boolean,
        bounds: SpatialSelectionBounds | null,
        operation?: "select" | "deselect",
    ) => void;
    extendSpatialSelection: (
        ids: number[],
        bounds: SpatialSelectionBounds | null,
    ) => void;
    endSelection: () => void;
    setSelectionMode: (enabled: boolean) => void;
    setDragging: (ids: number[]) => void;
    clear: () => void;
    reset: () => void;
}

const empty = () => ({
    selectedIds: [],
    draggingIds: [],
    selectionMode: false,
    selectionGesture: null,
    anchorId: null,
    selectionBounds: null,
});
const stores = new Map<string, UseBoundStore<StoreApi<GallerySelection>>>();

function createSelectionStore() {
    return create<GallerySelection>((set) => ({
        ...empty(),
        setDragging: (ids) => set({ draggingIds: ids }),
        toggle: (id) =>
            set((state) => ({
                selectedIds: state.selectedIds.includes(id)
                    ? state.selectedIds.filter((item) => item !== id)
                    : [...state.selectedIds, id],
            })),
        selectClick: (order, id, shift, additive) =>
            set((state) => {
                const next = beginSelection({
                    order,
                    target: id,
                    selected: state.selectedIds,
                    anchor: state.anchorId,
                    shift,
                    additive,
                });
                return {
                    selectedIds: next.selected,
                    anchorId: next.anchor,
                    selectionGesture: null,
                    selectionBounds: null,
                };
            }),
        selectMany: (ids) =>
            set({
                selectedIds: [...new Set(ids)],
                anchorId: ids.at(-1) ?? null,
                selectionGesture: null,
                selectionBounds: null,
            }),
        beginSelection: (order, id, shift, additive) =>
            set((state) => {
                const next = beginSelection({
                    order,
                    target: id,
                    selected: state.selectedIds,
                    anchor: state.anchorId,
                    shift,
                    additive,
                });
                return {
                    selectedIds: next.selected,
                    anchorId: next.anchor,
                    selectionGesture: next.gesture,
                };
            }),
        extendSelection: (id) =>
            set((state) =>
                state.selectionGesture
                    ? {
                          selectedIds: extendSelection(
                              state.selectionGesture,
                              id,
                          ),
                      }
                    : state,
            ),
        beginSpatialSelection: (
            ids,
            start,
            _target,
            additive,
            bounds,
            operation = "select",
        ) =>
            set((state) => {
                const base = additive ? state.selectedIds : [];
                const selectedIds =
                    operation === "select"
                        ? [...new Set([...base, ...ids])]
                        : base.filter((id) => !ids.includes(id));
                return {
                    selectedIds,
                    anchorId: start >= 0 ? start : null,
                    selectionBounds: bounds,
                    selectionGesture: {
                        order: ids,
                        start,
                        base,
                        mode: operation,
                        spatial: true,
                    },
                };
            }),
        extendSpatialSelection: (ids, bounds) =>
            set((state) => {
                const gesture = state.selectionGesture;
                if (!gesture?.spatial) return state;
                const range = [...new Set(ids)];
                return {
                    selectedIds:
                        gesture.mode === "select"
                            ? [...new Set([...gesture.base, ...range])]
                            : gesture.base.filter((id) => !range.includes(id)),
                    selectionBounds: bounds,
                };
            }),
        endSelection: () =>
            set({ selectionGesture: null, selectionBounds: null }),
        setSelectionMode: (enabled) =>
            set({
                selectionMode: enabled,
                selectionGesture: null,
                selectionBounds: null,
            }),
        clear: () =>
            set({
                selectedIds: [],
                draggingIds: [],
                selectionGesture: null,
                selectionBounds: null,
                anchorId: null,
            }),
        reset: () => set({ ...empty() }),
    }));
}

export function getGallerySelectionStore(scope: string) {
    let store = stores.get(scope);
    if (!store) {
        store = createSelectionStore();
        stores.set(scope, store);
    }
    return store;
}

export function useGallerySelection<T>(
    scope: string,
    selector: (state: GallerySelection) => T,
): T {
    return getGallerySelectionStore(scope)(selector);
}

export function getGallerySelectionState(scope: string): GallerySelection {
    return getGallerySelectionStore(scope).getState();
}

const useActiveScopeStore = create<{
    scope: string;
    setScope: (scope: string) => void;
}>((set) => ({ scope: "home", setScope: (scope) => set({ scope }) }));
export const useActiveGalleryScope = useActiveScopeStore;
export const setActiveGalleryScope = (scope: string) =>
    useActiveScopeStore.getState().setScope(scope);
