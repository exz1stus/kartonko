export type PaletteBase = "background" | "foreground" | "accent";
export type PaletteGroup = "primary" | "surface" | "tonal" | "controls" | "semantic";
export const derivedTokens = [
    "primary10", "primary20", "primary30", "primary40", "primary50",
    "surface10", "surface20", "surface30", "surface40", "surface50",
    "tonal0", "tonal10", "tonal20", "tonal30", "tonal40", "tonal50",
    "controlSurface", "controlForeground", "controlPlaceholder", "controlBorder",
    "mutedForeground", "selection", "selectionForeground", "primaryForeground",
] as const;
export type DerivedToken = (typeof derivedTokens)[number];
export type PaletteToken = PaletteBase | DerivedToken;
export type ImageTheme = {
    bases: Record<PaletteBase, string>;
    overrides: Partial<Record<DerivedToken, string>>;
    sourceColors: string[];
};
export type PaletteColors = Record<PaletteToken, string>;

export const paletteGroups: Record<PaletteGroup, DerivedToken[]> = {
    primary: ["primary10", "primary20", "primary30", "primary40", "primary50"],
    surface: ["surface10", "surface20", "surface30", "surface40", "surface50"],
    tonal: ["tonal0", "tonal10", "tonal20", "tonal30", "tonal40", "tonal50"],
    controls: ["controlSurface", "controlForeground", "controlPlaceholder", "controlBorder"],
    semantic: ["mutedForeground", "selection", "selectionForeground", "primaryForeground"],
};

export const paletteSources: Record<DerivedToken, PaletteToken[]> = {
    primary10: ["accent", "foreground"], primary20: ["accent", "foreground"],
    primary30: ["accent", "foreground"], primary40: ["accent", "foreground"],
    primary50: ["accent", "foreground"],
    surface10: ["background", "foreground"], surface20: ["background", "foreground"],
    surface30: ["background", "foreground"], surface40: ["background", "foreground"],
    surface50: ["background", "foreground"],
    tonal0: ["background", "accent"], tonal10: ["surface10", "accent"],
    tonal20: ["surface20", "accent"], tonal30: ["surface30", "accent"],
    tonal40: ["surface40", "accent"], tonal50: ["surface50", "accent"],
    controlSurface: ["background", "foreground"],
    controlForeground: ["controlSurface"],
    controlPlaceholder: ["controlForeground", "controlSurface"],
    controlBorder: ["controlForeground", "controlSurface"],
    mutedForeground: ["background", "foreground"],
    selection: ["accent"],
    selectionForeground: ["selection"],
    primaryForeground: ["accent"],
};

export function isHexColor(value: unknown): value is string {
    return typeof value === "string" && /^#[0-9a-f]{6}$/i.test(value);
}

function rgb(hex: string): number[] {
    return [1, 3, 5].map((index) => parseInt(hex.slice(index, index + 2), 16));
}

function hex(channels: number[]): string {
    return `#${channels.map((channel) => Math.round(channel).toString(16).padStart(2, "0")).join("")}`;
}

export function mixColor(first: string, second: string, secondWeight: number): string {
    const firstChannels = rgb(first);
    const secondChannels = rgb(second);
    return hex(firstChannels.map((channel, index) => channel * (1 - secondWeight) + secondChannels[index] * secondWeight));
}

export function contrastColor(color: string): string {
    const channels = rgb(color).map((value) => {
        const normalized = value / 255;
        return normalized <= 0.04045 ? normalized / 12.92 : ((normalized + 0.055) / 1.055) ** 2.4;
    });
    const luminance = channels[0] * 0.2126 + channels[1] * 0.7152 + channels[2] * 0.0722;
    return luminance > 0.179 ? "#171717" : "#ffffff";
}

function saturation(color: string): number {
    const channels = rgb(color);
    return Math.max(...channels) - Math.min(...channels);
}

export function themeFromImage(colors: string[], dark: boolean): ImageTheme {
    const sourceColors = colors.filter(isHexColor);
    const dominant = sourceColors[0] ?? "#4b0082";
    const vibrant = [...sourceColors].sort((first, second) => saturation(second) - saturation(first))[0] ?? dominant;
    const background = mixColor(dominant, dark ? "#101014" : "#ffffff", dark ? 0.76 : 0.86);
    const contrastTarget = dark ? "#ffffff" : "#111111";
    const foregroundSource = [...sourceColors].sort((first, second) =>
        Math.abs(rgb(second).reduce((sum, value) => sum + value, 0) - rgb(background).reduce((sum, value) => sum + value, 0))
        - Math.abs(rgb(first).reduce((sum, value) => sum + value, 0) - rgb(background).reduce((sum, value) => sum + value, 0)),
    )[0] ?? dominant;
    return {
        bases: {
            background,
            foreground: mixColor(foregroundSource, contrastTarget, 0.82),
            accent: mixColor(vibrant, dark ? "#ffffff" : "#111111", dark ? 0.28 : 0.16),
        },
        overrides: {},
        sourceColors,
    };
}

export function derivePalette(theme: ImageTheme, dark: boolean): PaletteColors {
    const { background, foreground, accent } = theme.bases;
    const colors = { ...theme.bases } as PaletteColors;
    const set = (token: DerivedToken, derived: string) => {
        colors[token] = theme.overrides[token] ?? derived;
    };
    (["primary10", "primary20", "primary30", "primary40", "primary50"] as const)
        .forEach((token, index) => set(token, mixColor(accent, foreground, (index + 1) * 0.12)));
    (["surface10", "surface20", "surface30", "surface40", "surface50"] as const)
        .forEach((token, index) => set(token, mixColor(background, foreground, [0.08, 0.16, 0.25, 0.35, 0.46][index])));
    set("tonal0", mixColor(background, accent, 0.08));
    (["tonal10", "tonal20", "tonal30", "tonal40", "tonal50"] as const)
        .forEach((token, index) => set(token, mixColor(colors[`surface${(index + 1) * 10}` as DerivedToken], accent, (index + 1) * 0.04 + 0.08)));
    set("controlSurface", dark
        ? mixColor(mixColor(colors.surface10, background, 0.72), accent, 0.08)
        : mixColor(mixColor(colors.surface10, colors.tonal10, 0.6), accent, 0.08));
    set("controlForeground", foreground);
    set("controlPlaceholder", mixColor(colors.controlForeground, colors.controlSurface, 0.42));
    set("controlBorder", mixColor(colors.controlSurface, accent, 0.3));
    set("mutedForeground", mixColor(foreground, background, 0.44));
    set("selection", accent);
    set("selectionForeground", contrastColor(colors.selection));
    set("primaryForeground", contrastColor(accent));
    return colors;
}

export const paletteCssVariables: Record<PaletteToken, string> = {
    background: "--clr-surface-a0", foreground: "--foreground", accent: "--clr-primary-a0",
    primary10: "--clr-primary-a10", primary20: "--clr-primary-a20",
    primary30: "--clr-primary-a30", primary40: "--clr-primary-a40", primary50: "--clr-primary-a50",
    surface10: "--clr-surface-a10", surface20: "--clr-surface-a20",
    surface30: "--clr-surface-a30", surface40: "--clr-surface-a40", surface50: "--clr-surface-a50",
    tonal0: "--clr-surface-tonal-a0", tonal10: "--clr-surface-tonal-a10",
    tonal20: "--clr-surface-tonal-a20", tonal30: "--clr-surface-tonal-a30",
    tonal40: "--clr-surface-tonal-a40", tonal50: "--clr-surface-tonal-a50",
    controlSurface: "--control-surface", controlForeground: "--control-foreground",
    controlPlaceholder: "--control-placeholder", controlBorder: "--control-border",
    mutedForeground: "--muted-foreground", selection: "--app-accent-color",
    selectionForeground: "--app-accent-foreground", primaryForeground: "--primary-foreground",
};

export const paletteAliases = ["--card-foreground", "--popover-foreground", "--accent-foreground"] as const;
