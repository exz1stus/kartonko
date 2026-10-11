export interface SelectionGesture<T> {
    order: T[];
    start: T;
    base: T[];
    mode: "select" | "deselect";
    spatial?: boolean;
    visited?: T[];
}

export interface SpatialSelectionBounds { left: number; top: number; width: number; height: number }

export function getPointerSelectionBounds(startX: number, startY: number, endX: number, endY: number): SpatialSelectionBounds {
    const left = Math.min(startX, endX);
    const top = Math.min(startY, endY);
    return { left, top, width: Math.abs(endX - startX), height: Math.abs(endY - startY) };
}

export function getPointerSelectionIds(container: ParentNode | null, bounds: SpatialSelectionBounds, scrollLeft = 0, scrollTop = 0): number[] {
    if (!container) return [];
    const right = bounds.left + bounds.width;
    const bottom = bounds.top + bounds.height;
    return Array.from(container.querySelectorAll<HTMLElement>("[data-gallery-image-id]"))
        .filter((card) => {
            const rect = card.getBoundingClientRect();
            return rect.right + scrollLeft >= bounds.left && rect.left + scrollLeft <= right && rect.bottom + scrollTop >= bounds.top && rect.top + scrollTop <= bottom;
        })
        .map((card) => Number(card.dataset.galleryImageId));
}

export function trackGalleryMarquee(
    grid: HTMLElement,
    startX: number,
    startY: number,
    update: (ids: number[], bounds: SpatialSelectionBounds) => void,
    finish: () => void,
): () => void {
    let pointerX = startX;
    let pointerY = startY;
    let frame: number | null = null;
    let stopped = false;
    let scrollElement: HTMLElement | null = null;
    for (let parent = grid.parentElement; parent; parent = parent.parentElement) {
        const overflowY = getComputedStyle(parent).overflowY;
        if ((overflowY === "auto" || overflowY === "scroll") && parent.scrollHeight > parent.clientHeight) {
            scrollElement = parent;
            break;
        }
    }
    const initialScrollLeft = scrollElement?.scrollLeft ?? 0;
    const initialScrollTop = scrollElement?.scrollTop ?? 0;

    const refresh = () => {
        const scrollLeft = scrollElement?.scrollLeft ?? 0;
        const scrollTop = scrollElement?.scrollTop ?? 0;
        const contentBounds = getPointerSelectionBounds(
            startX + initialScrollLeft,
            startY + initialScrollTop,
            pointerX + scrollLeft,
            pointerY + scrollTop,
        );
        const visibleBounds = getPointerSelectionBounds(
            startX + initialScrollLeft - scrollLeft,
            startY + initialScrollTop - scrollTop,
            pointerX,
            pointerY,
        );
        const ids = getPointerSelectionIds(grid, contentBounds, scrollLeft, scrollTop);
        if (scrollElement) {
            const viewport = scrollElement.getBoundingClientRect();
            const left = Math.max(visibleBounds.left, viewport.left);
            const top = Math.max(visibleBounds.top, viewport.top);
            const right = Math.min(visibleBounds.left + visibleBounds.width, viewport.right);
            const bottom = Math.min(visibleBounds.top + visibleBounds.height, viewport.bottom);
            update(ids, { left, top, width: Math.max(0, right - left), height: Math.max(0, bottom - top) });
        } else update(ids, visibleBounds);
    };
    const onMove = (event: PointerEvent) => {
        pointerX = event.clientX;
        pointerY = event.clientY;
        refresh();
        if (frame === null) frame = requestAnimationFrame(scrollAtEdge);
    };
    const onScroll = () => refresh();
    const stop = () => {
        if (stopped) return;
        stopped = true;
        window.removeEventListener("pointermove", onMove);
        window.removeEventListener("pointerup", stop);
        scrollElement?.removeEventListener("scroll", onScroll);
        if (frame !== null) cancelAnimationFrame(frame);
        finish();
    };
    const scrollAtEdge = () => {
        frame = null;
        if (stopped || !scrollElement) return;
        const rect = scrollElement.getBoundingClientRect();
        const edgeSize = 56;
        let delta = 0;
        if (pointerY < rect.top + edgeSize && pointerY >= rect.top) {
            delta = -Math.max(2, Math.ceil((rect.top + edgeSize - pointerY) / 5));
        } else if (pointerY > rect.bottom - edgeSize && pointerY <= rect.bottom) {
            delta = Math.max(2, Math.ceil((pointerY - (rect.bottom - edgeSize)) / 5));
        }
        if (delta !== 0) {
            const previous = scrollElement.scrollTop;
            scrollElement.scrollTop += delta;
            if (scrollElement.scrollTop !== previous) refresh();
            if ((delta < 0 && scrollElement.scrollTop > 0) || (delta > 0 && scrollElement.scrollTop + scrollElement.clientHeight < scrollElement.scrollHeight)) {
                frame = requestAnimationFrame(scrollAtEdge);
            }
        }
    };

    window.addEventListener("pointermove", onMove);
    window.addEventListener("pointerup", stop, { once: true });
    scrollElement?.addEventListener("scroll", onScroll, { passive: true });
    return stop;
}

export function getSpatialSelectionBounds(container: ParentNode | null, start: number, end: number): SpatialSelectionBounds | null {
    if (!container) return null;
    const cards = Array.from(container.querySelectorAll<HTMLElement>("[data-gallery-image-id]"));
    const startCard = cards.find((card) => Number(card.dataset.galleryImageId) === start);
    const endCard = cards.find((card) => Number(card.dataset.galleryImageId) === end);
    if (!startCard || !endCard) return null;
    const a = startCard.getBoundingClientRect();
    const b = endCard.getBoundingClientRect();
    const left = Math.min(a.left, b.left);
    const right = Math.max(a.right, b.right);
    const top = Math.min(a.top, b.top);
    const bottom = Math.max(a.bottom, b.bottom);
    return { left, top, width: right - left, height: bottom - top };
}

export function getSpatialSelectionIds(container: ParentNode | null, start: number, end: number): number[] {
    if (!container) return [end];
    const bounds = getSpatialSelectionBounds(container, start, end);
    if (!bounds) return [end];
    const right = bounds.left + bounds.width;
    const bottom = bounds.top + bounds.height;
    const cards = Array.from(container.querySelectorAll<HTMLElement>("[data-gallery-image-id]"));
    return cards.filter((card) => {
        const rect = card.getBoundingClientRect();
        const horizontalHit = rect.right >= bounds.left && rect.left <= right;
        const verticalHit = rect.bottom >= bounds.top && rect.top <= bottom;
        return horizontalHit && verticalHit;
    }).map((card) => Number(card.dataset.galleryImageId));
}

export function selectionRange<T>(order: T[], start: T, end: T): T[] {
    const startIndex = order.indexOf(start);
    const endIndex = order.indexOf(end);
    if (startIndex < 0 || endIndex < 0) return [end];
    const first = Math.min(startIndex, endIndex);
    const last = Math.max(startIndex, endIndex);
    return order.slice(first, last + 1);
}

export function beginSelection<T>(args: {
    order: T[];
    target: T;
    selected: T[];
    anchor: T | null;
    shift: boolean;
    additive: boolean;
}): { selected: T[]; anchor: T; gesture: SelectionGesture<T> } {
    const { order, target, selected, anchor, shift, additive } = args;

    if (shift) {
        const start = anchor !== null && order.includes(anchor) ? anchor : target;
        const base = additive ? selected : [];
        const range = selectionRange(order, start, target);
        return {
            selected: [...new Set([...base, ...range])],
            anchor: start,
            gesture: { order, start, base, mode: "select" },
        };
    }

    const alreadySelected = selected.includes(target);
    if (additive) {
        const mode = alreadySelected ? "deselect" : "select";
        const base = alreadySelected ? selected.filter((id) => id !== target) : selected;
        return {
            selected: alreadySelected ? base : [...base, target],
            anchor: target,
            gesture: { order, start: target, base, mode },
        };
    }

    return {
        selected: [target],
        anchor: target,
        gesture: { order, start: target, base: [], mode: "select" },
    };
}

export function extendSelection<T>(gesture: SelectionGesture<T>, target: T): T[] {
    const range = selectionRange(gesture.order, gesture.start, target);
    return gesture.mode === "select"
        ? [...new Set([...gesture.base, ...range])]
        : gesture.base.filter((id) => !range.includes(id));
}
