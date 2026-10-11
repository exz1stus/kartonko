"use client";

import { Button } from "@/components/ui/button";
import { useTranslations } from "@/contexts/LocaleContext";
import { Check, Crop, Maximize2, X } from "lucide-react";
import { createPortal } from "react-dom";
import { type PointerEvent, useEffect, useRef, useState } from "react";

type Mode = "fit" | "crop";
type CropRect = { x: number; y: number; width: number; height: number };
type DragState = { mode: "draw" | "move" | "resize"; start: { x: number; y: number }; crop: CropRect; anchor?: { x: number; y: number }; signX?: number; signY?: number };
type Props = {
    label: string;
    editLabel?: string;
    aspectRatio: number;
    currentFile?: File | null;
    accept?: string;
    disabled?: boolean;
    outputType?: string;
    maxOutputDimension?: number;
    onSelect: (file: File, mode: Mode, sourceFile: File) => void | Promise<void>;
};

function initialCrop(width: number, height: number, ratio: number): CropRect {
    const cropWidth = width / height > ratio ? height * ratio : width;
    const cropHeight = cropWidth / ratio;
    return { x: (width - cropWidth) / 2, y: (height - cropHeight) / 2, width: cropWidth, height: cropHeight };
}

function metrics(frame: HTMLDivElement, imageWidth: number, imageHeight: number) {
    const width = frame.clientWidth;
    const height = frame.clientHeight;
    const scale = Math.min(width / imageWidth, height / imageHeight);
    const shownWidth = imageWidth * scale;
    const shownHeight = imageHeight * scale;
    return { width, height, scale, left: (width - shownWidth) / 2, top: (height - shownHeight) / 2 };
}

function cropImage(source: string, file: File, crop: CropRect, cropFailedMessage: string, outputType?: string, maxOutputDimension?: number): Promise<File> {
    return new Promise((resolve, reject) => {
        const image = new window.Image();
        image.onload = () => {
            const maxSize = maxOutputDimension ?? Math.max(crop.width, crop.height);
            const scale = Math.min(1, maxSize / Math.max(crop.width, crop.height));
            const canvas = document.createElement("canvas");
            canvas.width = Math.max(1, Math.round(crop.width * scale));
            canvas.height = Math.max(1, Math.round(crop.height * scale));
            const context = canvas.getContext("2d");
            if (!context) return reject(new Error(cropFailedMessage));
            context.drawImage(image, crop.x, crop.y, crop.width, crop.height, 0, 0, canvas.width, canvas.height);
            const type = outputType || file.type || "image/png";
            canvas.toBlob((blob) => {
                if (!blob) return reject(new Error(cropFailedMessage));
                const extension = blob.type === "image/jpeg" ? "jpg" : blob.type.split("/")[1] || "png";
                const name = file.name.replace(/\.[^.]+$/, "") + `-crop.${extension}`;
                resolve(new File([blob], name, { type: blob.type || type, lastModified: Date.now() }));
            }, type, type === "image/jpeg" ? 0.9 : undefined);
        };
        image.onerror = () => reject(new Error(cropFailedMessage));
        image.src = source;
    });
}

export default function ImageSelector({ label, editLabel, aspectRatio, currentFile, accept = "image/*", disabled = false, outputType, maxOutputDimension, onSelect }: Props) {
    const { t } = useTranslations();
    const [file, setFile] = useState<File | null>(null);
    const [source, setSource] = useState("");
    const [mode, setMode] = useState<Mode>("fit");
    const [imageDimensions, setImageDimensions] = useState({ width: 0, height: 0 });
    const [frameDimensions, setFrameDimensions] = useState({ width: 0, height: 0 });
    const [crop, setCrop] = useState<CropRect | null>(null);
    const [busy, setBusy] = useState(false);
    const [operationError, setOperationError] = useState("");
    const frameRef = useRef<HTMLDivElement>(null);
    const dragState = useRef<DragState | null>(null);
    const inputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        if (!source || !frameRef.current) return;
        const frame = frameRef.current;
        const observer = new ResizeObserver(() => setFrameDimensions({ width: frame.clientWidth, height: frame.clientHeight }));
        observer.observe(frame);
        setFrameDimensions({ width: frame.clientWidth, height: frame.clientHeight });
        return () => observer.disconnect();
    }, [source]);

    useEffect(() => () => {
        if (source) URL.revokeObjectURL(source);
    }, [source]);

    useEffect(() => {
        if (!source) return;
        const closeOnEscape = (event: KeyboardEvent) => {
            if (event.key === "Escape" && !busy) close();
        };
        window.addEventListener("keydown", closeOnEscape);
        return () => window.removeEventListener("keydown", closeOnEscape);
    }, [source, busy]);

    const close = () => {
        setFile(null);
        setSource("");
        setCrop(null);
        setMode("fit");
    };

    const selectFile = (selected?: File) => {
        if (!selected) return;
        if (!selected.type.startsWith("image/")) return;
        setFile(selected);
        setSource(URL.createObjectURL(selected));
        setMode("fit");
        setOperationError("");
    };

    const openSelectedFile = () => {
        if (currentFile) selectFile(currentFile);
    };

    const currentMetrics = frameRef.current && imageDimensions.width > 0
        ? (frameDimensions.width > 0 ? metrics(frameRef.current, imageDimensions.width, imageDimensions.height) : null)
        : null;

    const pointInImage = (event: PointerEvent<HTMLDivElement>) => {
        const frame = frameRef.current;
        if (!currentMetrics || !frame) return null;
        const rect = frame.getBoundingClientRect();
        return {
            x: Math.max(0, Math.min(imageDimensions.width, (event.clientX - rect.left - currentMetrics.left) / currentMetrics.scale)),
            y: Math.max(0, Math.min(imageDimensions.height, (event.clientY - rect.top - currentMetrics.top) / currentMetrics.scale)),
        };
    };

    const cropStyle = crop && currentMetrics ? {
        left: `${(currentMetrics.left + crop.x * currentMetrics.scale) / currentMetrics.width * 100}%`,
        top: `${(currentMetrics.top + crop.y * currentMetrics.scale) / currentMetrics.height * 100}%`,
        width: `${crop.width * currentMetrics.scale / currentMetrics.width * 100}%`,
        height: `${crop.height * currentMetrics.scale / currentMetrics.height * 100}%`,
    } : undefined;

    const finish = async () => {
        if (!file) return;
        setBusy(true);
        try {
            const result = mode === "crop" && crop && source
                ? await cropImage(source, file, crop, t("imageSelector.cropFailed"), outputType, maxOutputDimension)
                : file;
            await onSelect(result, mode, file);
            close();
        } catch (error) {
            setOperationError(error instanceof Error ? error.message : t("imageSelector.cropFailed"));
        } finally {
            setBusy(false);
        }
    };

    return <>
        <label className={`inline-flex items-center justify-center gap-2 bg-surface-10 hover:bg-surface-20 px-3 py-2 border rounded-lg text-sm cursor-pointer ${disabled ? "opacity-50 pointer-events-none" : ""}`}>
            {label}
            <input ref={inputRef} type="file" accept={accept} disabled={disabled} className="sr-only" onChange={(event) => { selectFile(event.currentTarget.files?.[0]); event.currentTarget.value = ""; }} />
        </label>
        {source && createPortal(<div className="z-[100] fixed inset-0 flex justify-center items-center bg-black/60 p-4" onMouseDown={(event) => { if (event.target === event.currentTarget && !busy) close(); }}>
            <section role="dialog" aria-modal="true" aria-labelledby="image-selector-title" className="bg-background shadow-2xl p-4 border rounded-xl w-full max-w-2xl text-foreground">
                <header className="flex justify-between items-center gap-3 mb-3">
                    <h2 id="image-selector-title" className="font-semibold text-lg">{t("imageSelector.title")}</h2>
            <Button variant="ghost" size="icon-sm" aria-label={t("imageSelector.close")} onClick={close} disabled={busy}><X /></Button>
                </header>
                <div className="flex gap-2 mb-3">
                    <button type="button" aria-pressed={mode === "fit"} onClick={() => setMode("fit")} className="inline-flex items-center gap-2 bg-surface-10 text-foreground px-3 py-2 border rounded-lg text-sm aria-pressed:bg-primary aria-pressed:text-primary-foreground aria-pressed:border-primary"><Maximize2 size={16} />{t("imageSelector.fit")}</button>
                    <button type="button" aria-pressed={mode === "crop"} onClick={() => setMode("crop")} className="inline-flex items-center gap-2 bg-surface-10 text-foreground px-3 py-2 border rounded-lg text-sm aria-pressed:bg-primary aria-pressed:text-primary-foreground aria-pressed:border-primary"><Crop size={16} />{t("imageSelector.crop")}</button>
                </div>
                <p className="mb-2 text-muted-foreground text-xs">{mode === "crop" ? t("imageSelector.cropHint") : t("imageSelector.fitHint")}</p>
                {operationError && <p role="alert" className="mb-2 text-destructive text-xs">{operationError}</p>}
                <div ref={frameRef} className="relative bg-black/80 rounded-lg w-full h-[min(55vh,420px)] overflow-hidden touch-none select-none" onPointerDown={(event) => {
                    if (mode !== "crop" || !currentMetrics) return;
                    const point = pointInImage(event);
                    if (!point) return;
                    dragState.current = { mode: "draw", start: point, crop: crop ?? initialCrop(imageDimensions.width, imageDimensions.height, aspectRatio) };
                    event.currentTarget.setPointerCapture(event.pointerId);
                }} onPointerMove={(event) => {
                    const drag = dragState.current;
                    const point = pointInImage(event);
                    if (!drag || !point) return;
                    if (drag.mode === "move") {
                        const x = Math.max(0, Math.min(imageDimensions.width - drag.crop.width, drag.crop.x + point.x - drag.start.x));
                        const y = Math.max(0, Math.min(imageDimensions.height - drag.crop.height, drag.crop.y + point.y - drag.start.y));
                        setCrop({ ...drag.crop, x, y });
                        return;
                    }
                    const anchor = drag.mode === "resize" ? drag.anchor! : drag.start;
                    const sx = drag.mode === "resize" ? drag.signX! : point.x >= anchor.x ? 1 : -1;
                    const sy = drag.mode === "resize" ? drag.signY! : point.y >= anchor.y ? 1 : -1;
                    const wantedWidth = Math.max(Math.abs(point.x - anchor.x), Math.abs(point.y - anchor.y) * aspectRatio);
                    const width = Math.min(wantedWidth, sx > 0 ? imageDimensions.width - anchor.x : anchor.x, (sy > 0 ? imageDimensions.height - anchor.y : anchor.y) * aspectRatio);
                    const height = width / aspectRatio;
                    if (width > 2 && height > 2) setCrop({ x: sx > 0 ? anchor.x : anchor.x - width, y: sy > 0 ? anchor.y : anchor.y - height, width, height });
                }} onPointerUp={(event) => {
                    dragState.current = null;
                    if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId);
                }}>
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={source} alt={t("imageSelector.previewAlt")} className="absolute inset-0 w-full h-full object-contain pointer-events-none" onLoad={(event) => {
                        const image = event.currentTarget;
                        setImageDimensions({ width: image.naturalWidth, height: image.naturalHeight });
                        setCrop(initialCrop(image.naturalWidth, image.naturalHeight, aspectRatio));
                    }} />
                    {mode === "crop" && cropStyle && <div className="absolute border-2 border-white shadow-[0_0_0_9999px_rgba(0,0,0,0.48)] cursor-move" style={cropStyle} onPointerDown={(event) => {
                        event.stopPropagation();
                        const point = pointInImage(event);
                        if (!point || !crop) return;
                        const handle = (event.target as HTMLElement).dataset.cropHandle;
                        if (handle) {
                            const left = handle.includes("w");
                            const top = handle.includes("n");
                            dragState.current = { mode: "resize", start: point, crop, anchor: { x: left ? crop.x + crop.width : crop.x, y: top ? crop.y + crop.height : crop.y }, signX: left ? -1 : 1, signY: top ? -1 : 1 };
                        } else {
                            dragState.current = { mode: "move", start: point, crop };
                        }
                        frameRef.current?.setPointerCapture(event.pointerId);
                    }}>
                        <span className="absolute right-0 bottom-0 bg-black/70 px-1 text-white text-[10px] pointer-events-none">{t("imageSelector.cropRegion")}</span>
                        {(["nw", "ne", "sw", "se"] as const).map((handle) => <span key={handle} data-crop-handle={handle} aria-label={t("imageSelector.resizeCrop")} className={`absolute z-10 bg-white border-2 border-primary rounded-full size-4 ${handle.includes("n") ? "top-[-7px]" : "bottom-[-7px]"} ${handle.includes("w") ? "left-[-7px]" : "right-[-7px]"} ${handle === "nw" || handle === "se" ? "cursor-nwse-resize" : "cursor-nesw-resize"}`} />)}
                    </div>}
                </div>
                <footer className="flex justify-end gap-2 mt-4">
                    <Button variant="outline" onClick={close} disabled={busy}>{t("imageSelector.cancel")}</Button>
                    <Button className="bg-primary text-primary-foreground hover:bg-primary/90" onClick={() => void finish()} disabled={busy || imageDimensions.width === 0}>{busy ? t("imageSelector.working") : <><Check />{t("imageSelector.useImage")}</>}</Button>
                </footer>
            </section>
        </div>, document.body)}
        {currentFile && <button type="button" disabled={disabled} onClick={openSelectedFile} className="ml-2 inline-flex items-center gap-1 rounded-lg border border-primary-30 bg-surface-0 px-3 py-2 text-sm font-medium text-foreground hover:bg-surface-10 disabled:opacity-50"><Crop size={15} />{editLabel ?? t("imageSelector.editSelected")}</button>}
    </>;
}
