"use client";

import { Button } from "@/components/ui/button";
import ImageSelector from "@/components/ImageSelector";
import ImagePaletteEditor from "@/components/Navbar/ImagePaletteEditor";
import { extractPalette } from "@/lib/image/extractPalette";
import { contrastColor, derivePalette, derivedTokens, isHexColor, paletteAliases, paletteCssVariables, themeFromImage, type ImageTheme, type PaletteToken } from "@/lib/theme/imageTheme";
import { OPEN_APPEARANCE_SETTINGS_EVENT } from "@/lib/settingsEvents";
import { useTranslations } from "@/contexts/LocaleContext";
import { useTheme } from "next-themes";
import { Check, Image as ImageIcon, Settings, X } from "lucide-react";
import { createPortal } from "react-dom";
import { useEffect, useRef, useState } from "react";

type PresetAccent = "purple" | "blue" | "green" | "rose" | "amber";
type Accent = PresetAccent | "custom" | "image";
type Preferences = {
    accent: Accent;
    customAccent: string;
    imageTheme: ImageTheme | null;
    backgroundImage: string;
    backgroundOpacity: number;
    surfaceOpacity: number;
    contentOpacity: number;
    backgroundMode: "fit" | "crop";
};
const STORAGE_KEY = "kartonko-settings";
const IMAGE_DB = "kartonko-settings-assets";
const IMAGE_STORE = "assets";
const accents: PresetAccent[] = ["purple", "blue", "green", "rose", "amber"];
const swatches: Record<PresetAccent, string> = {
    purple: "#c49be9",
    blue: "#2878d0",
    green: "#16845b",
    rose: "#d44772",
    amber: "#b66a0a",
};

function readImageTheme(value: unknown): ImageTheme | null {
    if (!value || typeof value !== "object") return null;
    const theme = value as Partial<ImageTheme>;
    if (!theme.bases || !isHexColor(theme.bases.background) || !isHexColor(theme.bases.foreground) || !isHexColor(theme.bases.accent)) return null;
    const overrides: ImageTheme["overrides"] = {};
    for (const token of derivedTokens) {
        const color = theme.overrides?.[token];
        if (isHexColor(color)) overrides[token] = color;
    }
    return {
        bases: theme.bases,
        overrides,
        sourceColors: Array.isArray(theme.sourceColors) ? theme.sourceColors.filter(isHexColor) : [],
    };
}

function readPreferences(): Preferences {
    try {
        const value = JSON.parse(localStorage.getItem(STORAGE_KEY) || "{}");
        const imageTheme = readImageTheme(value.imageTheme);
        return {
            accent: value.accent === "image" && imageTheme ? "image" : value.accent === "custom" || accents.includes(value.accent) ? value.accent : "purple",
            customAccent: typeof value.customAccent === "string" && /^#[0-9a-f]{6}$/i.test(value.customAccent)
                ? value.customAccent
                : "#c49be9",
            imageTheme,
            backgroundImage:
                typeof value.backgroundImage === "string"
                    ? value.backgroundImage
                    : "/apilocal/image/name/bg/raw",
            backgroundOpacity: Number.isFinite(value.backgroundOpacity)
                ? Math.min(100, Math.max(0, value.backgroundOpacity))
                : 25,
            surfaceOpacity: Number.isFinite(value.surfaceOpacity)
                ? Math.min(100, Math.max(55, value.surfaceOpacity))
                : 88,
            contentOpacity: Number.isFinite(value.contentOpacity)
                ? Math.min(100, Math.max(0, value.contentOpacity))
                : 78,
            backgroundMode: value.backgroundMode === "crop" ? "crop" : "fit",
        };
    } catch {
        return {
            accent: "purple",
            customAccent: "#c49be9",
            imageTheme: null,
            backgroundImage: "/apilocal/image/name/bg/raw",
            backgroundOpacity: 25,
            surfaceOpacity: 88,
            contentOpacity: 78,
            backgroundMode: "fit",
        };
    }
}

function openImageDB(): Promise<IDBDatabase> {
    return new Promise((resolve, reject) => {
        const request = indexedDB.open(IMAGE_DB, 1);
        request.onupgradeneeded = () =>
            request.result.createObjectStore(IMAGE_STORE);
        request.onsuccess = () => resolve(request.result);
        request.onerror = () => reject(request.error);
    });
}

async function storeBackgroundImage(file: File, sourceFile: File) {
    const db = await openImageDB();
    await new Promise<void>((resolve, reject) => {
        const transaction = db.transaction(IMAGE_STORE, "readwrite");
        transaction.objectStore(IMAGE_STORE).put(file, "background");
        transaction
            .objectStore(IMAGE_STORE)
            .put(sourceFile, "background-original");
        transaction.oncomplete = () => resolve();
        transaction.onerror = () => reject(transaction.error);
    });
    db.close();
}

async function loadBackgroundImage(): Promise<Blob | undefined> {
    const db = await openImageDB();
    const image = await new Promise<Blob | undefined>((resolve, reject) => {
        const request = db
            .transaction(IMAGE_STORE)
            .objectStore(IMAGE_STORE)
            .get("background");
        request.onsuccess = () => resolve(request.result as Blob | undefined);
        request.onerror = () => reject(request.error);
    });
    db.close();
    return image;
}

async function loadOriginalBackgroundImage(): Promise<File | undefined> {
    const image =
        (await loadAsset("background-original")) ??
        (await loadAsset("background"));
    if (!image) return undefined;
    const file = image as File;
    return new File([image], file.name || "background-image", {
        type: image.type || "image/png",
        lastModified: file.lastModified || Date.now(),
    });
}

async function loadAsset(key: string): Promise<Blob | undefined> {
    const db = await openImageDB();
    const image = await new Promise<Blob | undefined>((resolve, reject) => {
        const request = db
            .transaction(IMAGE_STORE)
            .objectStore(IMAGE_STORE)
            .get(key);
        request.onsuccess = () => resolve(request.result as Blob | undefined);
        request.onerror = () => reject(request.error);
    });
    db.close();
    return image;
}

async function clearBackgroundImage() {
    const db = await openImageDB();
    await new Promise<void>((resolve, reject) => {
        const transaction = db.transaction(IMAGE_STORE, "readwrite");
        transaction.objectStore(IMAGE_STORE).delete("background");
        transaction.objectStore(IMAGE_STORE).delete("background-original");
        transaction.oncomplete = () => resolve();
        transaction.onerror = () => reject(transaction.error);
    });
    db.close();
}

export default function SettingsModal() {
    const { theme, resolvedTheme, setTheme } = useTheme();
    const { language, setLanguage, t } = useTranslations();
    const [shown, setShown] = useState(false);
    const [activeTab, setActiveTab] = useState<"general" | "appearance">(
        "general",
    );
    const [preferences, setPreferences] = useState<Preferences>(() =>
        typeof window === "undefined"
            ? {
                  accent: "purple",
                  customAccent: "#c49be9",
                  imageTheme: null,
                  backgroundImage: "/apilocal/image/name/bg/raw",
                  backgroundOpacity: 25,
                  surfaceOpacity: 88,
                  contentOpacity: 78,
                  backgroundMode: "fit",
              }
            : readPreferences(),
    );
    const [localImageUrl, setLocalImageUrl] = useState<string | null>(null);
    const [localOriginalFile, setLocalOriginalFile] = useState<File | null>(
        null,
    );
    const [imageError, setImageError] = useState("");
    const [paletteLoading, setPaletteLoading] = useState(false);
    const [paletteError, setPaletteError] = useState("");
    const imageUrlRef = useRef<string | null>(null);
    const paletteRequestRef = useRef(0);

    useEffect(() => {
        const openAppearance = () => {
            setActiveTab("appearance");
            setShown(true);
        };
        window.addEventListener(OPEN_APPEARANCE_SETTINGS_EVENT, openAppearance);
        return () => window.removeEventListener(OPEN_APPEARANCE_SETTINGS_EVENT, openAppearance);
    }, []);

    useEffect(() => {
        if (preferences.backgroundImage !== "local" || imageUrlRef.current)
            return;
        let cancelled = false;
        void Promise.all([loadBackgroundImage(), loadOriginalBackgroundImage()])
            .then(([image, original]) => {
                if (!image || cancelled) return;
                const url = URL.createObjectURL(image);
                imageUrlRef.current = url;
                setLocalImageUrl(url);
                setLocalOriginalFile(original ?? null);
            })
            .catch(() => setImageError(t("settings.background.loadError")));
        return () => {
            cancelled = true;
        };
    }, [preferences.backgroundImage, t]);

    useEffect(() => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(preferences));
        const root = document.documentElement;
        root.dataset.accent = preferences.accent === "image" ? "custom" : preferences.accent;
        const palette = preferences.accent === "image" && preferences.imageTheme
            ? derivePalette(preferences.imageTheme, resolvedTheme === "dark")
            : null;
        for (const [token, variable] of Object.entries(paletteCssVariables)) {
            if (palette) root.style.setProperty(variable, palette[token as PaletteToken]);
            else root.style.removeProperty(variable);
        }
        for (const variable of paletteAliases) {
            if (palette) root.style.setProperty(variable, palette.foreground);
            else root.style.removeProperty(variable);
        }
        const accentColor = palette?.selection ?? (preferences.accent === "custom" || preferences.accent === "image"
            ? preferences.customAccent
            : preferences.accent === "purple" && resolvedTheme !== "dark"
                ? "#7650a7"
                : swatches[preferences.accent]);
        root.style.setProperty("--app-accent-color", accentColor);
        root.style.setProperty("--app-accent-foreground", palette?.selectionForeground ?? contrastColor(accentColor));
        root.style.setProperty("--app-surface-opacity", `${preferences.surfaceOpacity}%`);
        root.style.setProperty("--app-content-opacity", `${preferences.contentOpacity}%`);
        const image =
            preferences.backgroundImage === "local"
                ? localImageUrl
                : preferences.backgroundImage;
        document.documentElement.style.setProperty(
            "--app-background-image",
            image ? `url(${JSON.stringify(image)})` : "none",
        );
        document.documentElement.style.setProperty(
            "--app-background-opacity",
            String(preferences.backgroundOpacity / 100),
        );
        document.documentElement.style.setProperty(
            "--app-background-size",
            preferences.backgroundMode === "fit" ? "contain" : "cover",
        );
    }, [preferences, localImageUrl, resolvedTheme]);

    useEffect(
        () => () => {
            if (imageUrlRef.current) URL.revokeObjectURL(imageUrlRef.current);
        },
        [],
    );

    useEffect(() => {
        if (!shown) return;
        const closeOnEscape = (event: KeyboardEvent) => {
            if (event.key === "Escape") setShown(false);
        };
        window.addEventListener("keydown", closeOnEscape);
        return () => window.removeEventListener("keydown", closeOnEscape);
    }, [shown]);

    const updatePreference = <K extends keyof Preferences>(
        key: K,
        value: Preferences[K],
    ) => {
        setPreferences((current) => ({ ...current, [key]: value }));
    };

    const clearImagePalette = () => {
        paletteRequestRef.current += 1;
        setPaletteError("");
        setPaletteLoading(false);
    };

    const generateImagePalette = async () => {
        const request = ++paletteRequestRef.current;
        setPaletteLoading(true);
        setPaletteError("");
        try {
            const source = preferences.backgroundImage === "local"
                ? await loadBackgroundImage()
                : preferences.backgroundImage.trim();
            if (!source) throw new Error("No background image");
            const colors = await extractPalette(source);
            if (colors.length === 0) throw new Error("No visible colors");
            if (request === paletteRequestRef.current) {
                setPreferences((current) => ({
                    ...current,
                    accent: "image",
                    imageTheme: themeFromImage(colors, resolvedTheme === "dark"),
                }));
            }
        } catch {
            if (request === paletteRequestRef.current) setPaletteError(t("settings.colors.imageError"));
        } finally {
            if (request === paletteRequestRef.current) setPaletteLoading(false);
        }
    };

    const removeBackgroundImage = async () => {
        clearImagePalette();
        if (imageUrlRef.current) URL.revokeObjectURL(imageUrlRef.current);
        imageUrlRef.current = null;
        setLocalImageUrl(null);
        setLocalOriginalFile(null);
        updatePreference("backgroundImage", "");
        try {
            await clearBackgroundImage();
        } catch {
            /* No stored file to remove. */
        }
    };

    const selectBackgroundImage = async (
        file: File,
        mode: "fit" | "crop",
        sourceFile: File,
    ) => {
        clearImagePalette();
        try {
            await storeBackgroundImage(file, sourceFile);
            const url = URL.createObjectURL(file);
            if (imageUrlRef.current) URL.revokeObjectURL(imageUrlRef.current);
            imageUrlRef.current = url;
            setLocalImageUrl(url);
            setLocalOriginalFile(sourceFile);
            setImageError("");
            setPreferences((current) => ({
                ...current,
                backgroundImage: "local",
                backgroundMode: mode,
            }));
        } catch {
            setImageError(t("settings.background.saveError"));
        }
    };

    return (
        <>
            <Button
                variant="ghost"
                size="icon"
                aria-label={t("settings.title")}
                onClick={() => setShown(true)}
            >
                <Settings aria-hidden="true" />
            </Button>
            {shown &&
                createPortal(
                    <div
                        className="z-50 fixed inset-0 flex justify-center items-center bg-black/55 p-4"
                        onMouseDown={(event) => {
                            if (event.target === event.currentTarget)
                                setShown(false);
                        }}
                    >
                        <section
                            role="dialog"
                            aria-modal="true"
                            aria-labelledby="settings-title"
                            className={`interface-chrome bg-background shadow-2xl p-5 border rounded-xl w-full max-h-[90vh] overflow-y-auto text-foreground ${activeTab === "appearance" && preferences.accent === "image" ? "max-w-2xl" : "max-w-md"}`}
                        >
                            <header className="flex justify-between items-center mb-5">
                                <h2
                                    id="settings-title"
                                    className="font-semibold text-xl"
                                >
                                    {t("settings.title")}
                                </h2>
                                <Button
                                    variant="ghost"
                                    size="icon-sm"
                                    aria-label={t("settings.close")}
                                    onClick={() => setShown(false)}
                                >
                                    <X />
                                </Button>
                            </header>
                            <div className="space-y-5">
                                <div
                                    role="tablist"
                                    aria-label={t("settings.title")}
                                    className="gap-1 grid grid-cols-2 bg-surface-10 p-1 rounded-lg"
                                >
                                    {(["general", "appearance"] as const).map(
                                        (tab) => (
                                            <button
                                                key={tab}
                                                type="button"
                                                role="tab"
                                                aria-selected={
                                                    activeTab === tab
                                                }
                                                onClick={() =>
                                                    setActiveTab(tab)
                                                }
                                                className="aria-selected:bg-surface-0 aria-selected:shadow-sm px-3 py-2 rounded-md text-sm"
                                            >
                                                {tab === "general"
                                                    ? t("settings.general")
                                                    : t("settings.appearance")}
                                            </button>
                                        ),
                                    )}
                                </div>
                                {activeTab === "appearance" && (
                                    <>
                                        <fieldset>
                                            <legend className="mb-2 font-medium text-sm">
                                                {t("settings.theme")}
                                            </legend>
                                            <div className="gap-2 grid grid-cols-3">
                                                {(
                                                    [
                                                        "system",
                                                        "light",
                                                        "dark",
                                                    ] as const
                                                ).map((option) => (
                                                    <button
                                                        key={option}
                                                        type="button"
                                                        aria-pressed={
                                                            (theme ||
                                                                "system") ===
                                                            option
                                                        }
                                                        onClick={() =>
                                                            setTheme(option)
                                                        }
                                                        className="flex justify-between items-center bg-surface-10 hover:bg-surface-20 px-3 py-2 border aria-pressed:border-primary-30 rounded-lg aria-pressed:ring-1 aria-pressed:ring-primary-30 text-sm capitalize"
                                                    >
                                                        {t(
                                                            `settings.theme.${option}` as Parameters<
                                                                typeof t
                                                            >[0],
                                                        )}{" "}
                                                        {(theme || "system") ===
                                                            option && (
                                                            <Check className="size-4" />
                                                        )}
                                                    </button>
                                                ))}
                                            </div>
                                        </fieldset>
                                        <fieldset className="space-y-2">
                                            <legend className="mb-2 font-medium text-sm">
                                                {t("settings.colors")}
                                            </legend>
                                            <div className="flex flex-wrap gap-2">
                                                {accents.map((accent) => (
                                                    <button
                                                        key={accent}
                                                        type="button"
                                                        aria-label={t(
                                                            `settings.colors.${accent}` as Parameters<
                                                                typeof t
                                                            >[0],
                                                        )}
                                                        aria-pressed={
                                                            preferences.accent ===
                                                            accent
                                                        }
                                                        onClick={() =>
                                                            updatePreference(
                                                                "accent",
                                                                accent,
                                                            )
                                                        }
                                                        className="flex justify-center items-center border-2 rounded-full aria-pressed:ring-2 aria-pressed:ring-offset-2 size-9"
                                                        style={{
                                                            backgroundColor:
                                                                swatches[
                                                                    accent
                                                                ],
                                                            borderColor:
                                                                preferences.accent ===
                                                                accent
                                                                    ? "var(--foreground)"
                                                                    : "transparent",
                                                        }}
                                                    >
                                                        {preferences.accent ===
                                                            accent && (
                                                            <Check className="size-4" style={{ color: contrastColor(swatches[accent]) }} />
                                                        )}
                                                    </button>
                                                ))}
                                                <label
                                                    className="relative flex size-9 cursor-pointer items-center justify-center rounded-full border-2 focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-primary-0"
                                                    style={{
                                                        backgroundColor: preferences.customAccent,
                                                        borderColor: preferences.accent === "custom" ? "var(--foreground)" : "transparent",
                                                        color: contrastColor(preferences.customAccent),
                                                    }}
                                                >
                                                    <input
                                                        type="color"
                                                        value={preferences.customAccent}
                                                        aria-label={t("settings.colors.custom")}
                                                        className="absolute inset-0 size-full cursor-pointer opacity-0"
                                                        onClick={() => updatePreference("accent", "custom")}
                                                        onChange={(event) => setPreferences((current) => ({
                                                            ...current,
                                                            accent: "custom",
                                                            customAccent: event.target.value,
                                                        }))}
                                                    />
                                                    {preferences.accent === "custom" && <Check aria-hidden="true" className="pointer-events-none size-4" />}
                                                </label>
                                                {preferences.imageTheme && (
                                                    <button
                                                        type="button"
                                                        aria-label={t("settings.palette.title")}
                                                        aria-pressed={preferences.accent === "image"}
                                                        onClick={() => updatePreference("accent", "image")}
                                                        className="flex size-9 items-center justify-center rounded-full border-2 aria-pressed:ring-2 aria-pressed:ring-offset-2"
                                                        style={{
                                                            background: `linear-gradient(135deg, ${preferences.imageTheme.bases.background} 48%, ${preferences.imageTheme.bases.accent} 52%)`,
                                                            borderColor: preferences.accent === "image" ? "var(--foreground)" : "transparent",
                                                            color: contrastColor(preferences.imageTheme.bases.background),
                                                        }}
                                                    >
                                                        <ImageIcon aria-hidden="true" className="size-4" />
                                                    </button>
                                                )}
                                                <span className="self-center font-mono text-xs text-muted-foreground">{preferences.accent === "custom" ? preferences.customAccent.toUpperCase() : ""}</span>
                                            </div>
                                            <button
                                                type="button"
                                                disabled={!preferences.backgroundImage.trim() || paletteLoading}
                                                onClick={() => void generateImagePalette()}
                                                className="rounded-lg border border-surface-30 bg-surface-10 px-3 py-2 text-sm hover:bg-surface-20 disabled:cursor-not-allowed disabled:opacity-50"
                                            >
                                                {paletteLoading ? t("settings.colors.readingImage") : t("settings.colors.fromBackground")}
                                            </button>
                                            {preferences.accent === "image" && preferences.imageTheme && (
                                                <ImagePaletteEditor
                                                    theme={preferences.imageTheme}
                                                    dark={resolvedTheme === "dark"}
                                                    onChange={(imageTheme) => updatePreference("imageTheme", imageTheme)}
                                                />
                                            )}
                                            {paletteError && <p role="alert" className="text-red-500 text-sm">{paletteError}</p>}
                                        </fieldset>
                                        <fieldset className="space-y-3">
                                            <legend className="mb-2 font-medium text-sm">
                                                {t("settings.background")}
                                            </legend>
                                            <label className="block text-sm">
                                                <span className="block mb-1">
                                                    {t(
                                                        "settings.background.image",
                                                    )}
                                                </span>
                                                <input
                                                    type="text"
                                                    value={
                                                        preferences.backgroundImage ===
                                                        "local"
                                                            ? ""
                                                            : preferences.backgroundImage
                                                    }
                                                    onChange={(event) => {
                                                        clearImagePalette();
                                                        updatePreference("backgroundImage", event.target.value);
                                                    }}
                                                    placeholder={
                                                        preferences.backgroundImage ===
                                                        "local"
                                                            ? t(
                                                                  "settings.background.localImage",
                                                              )
                                                            : t(
                                                                  "settings.background.imagePlaceholder",
                                                              )
                                                    }
                                                    className="bg-surface-10 px-3 py-2 border rounded-lg w-full font-normal"
                                                />
                                            </label>
                                            <div className="flex flex-wrap items-center gap-2">
                                                <ImageSelector
                                                    label={t(
                                                        "settings.background.browse",
                                                    )}
                                                    editLabel={t(
                                                        "imageSelector.adjustSelected",
                                                    )}
                                                    currentFile={
                                                        preferences.backgroundImage ===
                                                        "local"
                                                            ? localOriginalFile
                                                            : null
                                                    }
                                                    aspectRatio={16 / 9}
                                                    onSelect={
                                                        selectBackgroundImage
                                                    }
                                                />
                                                <label className="inline-flex items-center gap-2 text-sm">
                                                    <span>
                                                        {t(
                                                            "settings.background.mode",
                                                        )}
                                                    </span>
                                                    <select
                                                        aria-label={t(
                                                            "settings.background.mode",
                                                        )}
                                                        value={
                                                            preferences.backgroundMode
                                                        }
                                                        onChange={(event) =>
                                                            updatePreference(
                                                                "backgroundMode",
                                                                event.target
                                                                    .value as
                                                                    | "fit"
                                                                    | "crop",
                                                            )
                                                        }
                                                        className="bg-surface-10 px-3 py-2 border rounded-lg"
                                                    >
                                                        <option value="fit">
                                                            {t(
                                                                "imageSelector.fit",
                                                            )}
                                                        </option>
                                                        <option value="crop">
                                                            {t(
                                                                "imageSelector.crop",
                                                            )}
                                                        </option>
                                                    </select>
                                                </label>
                                                {(preferences.backgroundImage ||
                                                    localImageUrl) && (
                                                    <button
                                                        type="button"
                                                        onClick={() =>
                                                            void removeBackgroundImage()
                                                        }
                                                        className="hover:bg-surface-10 px-3 py-2 rounded-lg text-muted-foreground text-sm"
                                                    >
                                                        {t(
                                                            "settings.background.remove",
                                                        )}
                                                    </button>
                                                )}
                                            </div>
                                            {imageError && (
                                                <p
                                                    role="alert"
                                                    className="text-destructive text-xs"
                                                >
                                                    {imageError}
                                                </p>
                                            )}
                                            <label className="block text-sm">
                                                <span className="flex justify-between mb-1">
                                                    <span>
                                                        {t(
                                                            "settings.background.opacity",
                                                        )}
                                                    </span>
                                                    <span className="tabular-nums text-muted-foreground">
                                                        {
                                                            preferences.backgroundOpacity
                                                        }
                                                        %
                                                    </span>
                                                </span>
                                                <input
                                                    type="range"
                                                    min="0"
                                                    max="100"
                                                    step="1"
                                                    value={
                                                        preferences.backgroundOpacity
                                                    }
                                                    onChange={(event) =>
                                                        updatePreference(
                                                            "backgroundOpacity",
                                                            Number(
                                                                event.target
                                                                    .value,
                                                            ),
                                                        )
                                                    }
                                                    className="w-full accent-primary-20"
                                                />
                                            </label>
                                            <label className="block text-sm">
                                                <span className="flex justify-between mb-1">
                                                    <span>
                                                        {t(
                                                            "settings.surfaceOpacity",
                                                        )}
                                                    </span>
                                                    <span className="tabular-nums text-muted-foreground">
                                                        {
                                                            preferences.surfaceOpacity
                                                        }
                                                        %
                                                    </span>
                                                </span>
                                                <input
                                                    type="range"
                                                    min="55"
                                                    max="100"
                                                    step="1"
                                                    value={
                                                        preferences.surfaceOpacity
                                                    }
                                                    onChange={(event) =>
                                                        updatePreference(
                                                            "surfaceOpacity",
                                                            Number(
                                                                event.target
                                                                    .value,
                                                            ),
                                                        )
                                                    }
                                                    className="w-full accent-primary-20"
                                                />
                                                <span className="block mt-1 text-muted-foreground text-xs">
                                                    {t(
                                                        "settings.surfaceOpacityHelp",
                                                    )}
                                                </span>
                                            </label>
                                            <label className="block text-sm">
                                                <span className="flex justify-between mb-1">
                                                    <span>{t("settings.contentOpacity")}</span>
                                                    <span className="tabular-nums text-muted-foreground">{preferences.contentOpacity}%</span>
                                                </span>
                                                <input
                                                    type="range"
                                                    min="0"
                                                    max="100"
                                                    step="1"
                                                    value={preferences.contentOpacity}
                                                    onChange={(event) => updatePreference("contentOpacity", Number(event.target.value))}
                                                    className="w-full accent-primary-20"
                                                />
                                                <span className="block mt-1 text-muted-foreground text-xs">{t("settings.contentOpacityHelp")}</span>
                                            </label>
                                        </fieldset>
                                    </>
                                )}
                                {activeTab === "general" && (
                                    <label className="block font-medium text-sm">
                                        <span className="block mb-2">
                                            {t("settings.language")}
                                        </span>
                                        <select
                                            value={language}
                                            onChange={(event) =>
                                                setLanguage(
                                                    event.target.value as
                                                        | "en"
                                                        | "mk"
                                                        | "rue",
                                                )
                                            }
                                            className="bg-surface-10 px-3 py-2 border rounded-lg w-full font-normal"
                                        >
                                            <option value="en">
                                                {t("settings.language.english")}
                                            </option>
                                            <option value="mk">
                                                {t(
                                                    "settings.language.macedonian",
                                                )}
                                            </option>
                                            <option value="rue">
                                                {t(
                                                    "settings.language.carpathoRusyn",
                                                )}
                                            </option>
                                        </select>
                                        <span className="block mt-1 font-normal text-muted-foreground text-xs">
                                            {t("settings.language.help")}
                                        </span>
                                    </label>
                                )}
                            </div>
                        </section>
                    </div>,
                    document.body,
                )}
        </>
    );
}
