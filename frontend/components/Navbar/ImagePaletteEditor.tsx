"use client";

import { useState } from "react";
import { RotateCcw } from "lucide-react";
import { useTranslations } from "@/contexts/LocaleContext";
import {
    contrastColor,
    derivePalette,
    paletteGroups,
    paletteSources,
    type DerivedToken,
    type ImageTheme,
    type PaletteBase,
    type PaletteGroup,
    type PaletteToken,
} from "@/lib/theme/imageTheme";

const bases: PaletteBase[] = ["background", "foreground", "accent"];
const groups: PaletteGroup[] = ["primary", "surface", "tonal", "controls", "semantic"];

export default function ImagePaletteEditor({ theme, dark, onChange }: { theme: ImageTheme; dark: boolean; onChange: (theme: ImageTheme) => void }) {
    const { t } = useTranslations();
    const [selectedToken, setSelectedToken] = useState<PaletteToken>("accent");
    const colors = derivePalette(theme, dark);

    const label = (token: PaletteToken): string => {
        if (token === "background") return t("settings.palette.background");
        if (token === "foreground") return t("settings.palette.foreground");
        if (token === "accent") return t("settings.palette.accent");
        if (token === "primaryForeground") return t("settings.palette.primaryForeground");
        if (token.startsWith("primary")) return `${t("settings.palette.accent")} ${token.slice(7)}`;
        if (token.startsWith("surface")) return `${t("settings.palette.surface")} ${token.slice(7)}`;
        if (token.startsWith("tonal")) return `${t("settings.palette.tonal")} ${token.slice(5)}`;
        switch (token) {
            case "controlSurface": return t("settings.palette.controlSurface");
            case "controlForeground": return t("settings.palette.controlForeground");
            case "controlPlaceholder": return t("settings.palette.controlPlaceholder");
            case "controlBorder": return t("settings.palette.controlBorder");
            case "mutedForeground": return t("settings.palette.mutedForeground");
            case "selection": return t("settings.palette.selection");
            case "selectionForeground": return t("settings.palette.selectionForeground");
        }
        return token;
    };

    const groupLabel = (group: PaletteGroup): string => {
        switch (group) {
            case "primary": return t("settings.palette.groupPrimary");
            case "surface": return t("settings.palette.groupSurface");
            case "tonal": return t("settings.palette.groupTonal");
            case "controls": return t("settings.palette.groupControls");
            case "semantic": return t("settings.palette.groupSemantic");
        }
    };

    const setColor = (token: PaletteToken, color: string) => {
        if (bases.includes(token as PaletteBase)) {
            onChange({ ...theme, bases: { ...theme.bases, [token]: color } });
        } else {
            onChange({ ...theme, overrides: { ...theme.overrides, [token]: color } });
        }
    };

    const resetColor = (token: DerivedToken) => {
        const overrides = { ...theme.overrides };
        delete overrides[token];
        onChange({ ...theme, overrides });
    };

    const renderColor = (token: PaletteToken) => {
        const isBase = bases.includes(token as PaletteBase);
        const override = !isBase && theme.overrides[token as DerivedToken] !== undefined;
        const description = isBase
            ? t(`settings.palette.affects.${token}` as "settings.palette.affects.background" | "settings.palette.affects.foreground" | "settings.palette.affects.accent")
            : override
              ? t("settings.palette.manual")
              : t("settings.palette.derivedFrom", {
                    sources: paletteSources[token as DerivedToken].map(label).join(" + "),
                });
        return (
            <div key={token} className={`flex min-w-0 items-center gap-2 rounded-lg border p-2 ${selectedToken === token ? "border-primary-0 bg-surface-tonal-0" : "border-surface-20 bg-surface-10"}`}>
                <input
                    type="color"
                    value={colors[token]}
                    aria-label={label(token)}
                    title={label(token)}
                    onFocus={() => setSelectedToken(token)}
                    onChange={(event) => setColor(token, event.target.value)}
                    className="size-9 shrink-0 cursor-pointer rounded-md border-0 bg-transparent p-0"
                />
                <button type="button" onClick={() => setSelectedToken(token)} aria-pressed={selectedToken === token} className="min-w-0 flex-1 text-left">
                    <span className="block truncate text-sm font-medium">{label(token)}</span>
                    <span className="block truncate text-xs text-muted-foreground" title={description}>{description}</span>
                </button>
                <span className="shrink-0 font-mono text-[11px] text-muted-foreground">{colors[token].toUpperCase()}</span>
                {override && <button type="button" onClick={() => resetColor(token as DerivedToken)} aria-label={t("settings.palette.reset", { name: label(token) })} title={t("settings.palette.reset", { name: label(token) })} className="rounded-md p-1 hover:bg-surface-20"><RotateCcw aria-hidden="true" size={15} /></button>}
            </div>
        );
    };

    return (
        <section aria-label={t("settings.palette.title")} className="space-y-3 rounded-xl border border-surface-20 bg-surface-0/80 p-3">
            <div>
                <h3 className="font-semibold">{t("settings.palette.title")}</h3>
                <p className="text-sm text-muted-foreground">{t("settings.palette.help")}</p>
            </div>
            <div className="grid gap-2 sm:grid-cols-3">{bases.map(renderColor)}</div>
            <div>
                <p className="mb-2 text-xs text-muted-foreground">{t("settings.palette.chooseFor", { name: label(selectedToken) })}</p>
                <div className="flex flex-wrap gap-2" role="group" aria-label={t("settings.palette.imageColors")}>
                    {theme.sourceColors.map((color) => (
                        <button
                            key={color}
                            type="button"
                            onClick={() => setColor(selectedToken, color)}
                            aria-label={t("settings.palette.useImageColor", { color, name: label(selectedToken) })}
                            className="flex size-9 items-center justify-center rounded-full border-2 border-surface-30 hover:ring-2 hover:ring-primary-0"
                            style={{ backgroundColor: color, color: contrastColor(color) }}
                        >
                            {colors[selectedToken].toLowerCase() === color.toLowerCase() && <span aria-hidden="true">✓</span>}
                        </button>
                    ))}
                </div>
            </div>
            {groups.map((group) => (
                <details key={group} className="rounded-lg border border-surface-20 bg-surface-0/80">
                    <summary className="cursor-pointer px-3 py-2 text-sm font-medium">{groupLabel(group)} <span className="text-muted-foreground">({paletteGroups[group].length})</span></summary>
                    <div className="grid gap-2 border-t border-surface-20 p-2 sm:grid-cols-2">
                        {paletteGroups[group].map(renderColor)}
                    </div>
                </details>
            ))}
        </section>
    );
}
