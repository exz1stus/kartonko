const cyrillicToLatinMap: Record<string, string> = {
    а: "a",
    б: "b",
    в: "v",
    г: "g",
    д: "d",
    е: "e",
    ё: "e",
    ж: "zh",
    з: "z",
    и: "i",
    й: "y",
    к: "k",
    л: "l",
    м: "m",
    н: "n",
    о: "o",
    п: "p",
    р: "r",
    с: "s",
    т: "t",
    у: "u",
    ф: "f",
    х: "h",
    ц: "ts",
    ч: "ch",
    ш: "sh",
    щ: "sht",
    ъ: "a",
    ь: "",
    ю: "yu",
    я: "ya",

    є: "ye",
    ґ: "g",
    і: "i",
    ї: "i",
    ў: "w",
};

export function sanitizeName(value: string): string {
    return value
        .toLowerCase()
        .split("")
        .map((char) => cyrillicToLatinMap[char] ?? char)
        .join("")
        .replace(/\s+/g, "_")
        .replace(/[^a-z0-9_]/g, "")
        .replace(/^_+/, "")
        .replace(/_+/g, "_");
}
