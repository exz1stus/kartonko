type ColorBucket = {
    count: number;
    red: number;
    green: number;
    blue: number;
};

function toHex(red: number, green: number, blue: number): string {
    return `#${[red, green, blue]
        .map((channel) => Math.round(channel).toString(16).padStart(2, "0"))
        .join("")}`;
}

async function loadImage(source: string): Promise<HTMLImageElement> {
    const image = new Image();
    image.crossOrigin = "anonymous";
    image.src = source;
    await image.decode();
    return image;
}

export async function extractPalette(source: Blob | string): Promise<string[]> {
    const image = source instanceof Blob ? await createImageBitmap(source) : await loadImage(source);
    try {
        const canvas = document.createElement("canvas");
        canvas.width = 96;
        canvas.height = 96;
        const context = canvas.getContext("2d", { willReadFrequently: true });
        if (!context) throw new Error("Canvas is unavailable");
        context.drawImage(image, 0, 0, canvas.width, canvas.height);
        const pixels = context.getImageData(0, 0, canvas.width, canvas.height).data;
        const buckets = new Map<number, ColorBucket>();

        for (let index = 0; index < pixels.length; index += 4) {
            if (pixels[index + 3] < 128) continue;
            const red = pixels[index];
            const green = pixels[index + 1];
            const blue = pixels[index + 2];
            const key = (red >> 4) << 8 | (green >> 4) << 4 | (blue >> 4);
            const bucket = buckets.get(key) ?? { count: 0, red: 0, green: 0, blue: 0 };
            bucket.count += 1;
            bucket.red += red;
            bucket.green += green;
            bucket.blue += blue;
            buckets.set(key, bucket);
        }

        const candidates = [...buckets.values()]
            .map((bucket) => {
                const color = [bucket.red, bucket.green, bucket.blue].map((sum) => sum / bucket.count);
                const saturation = (Math.max(...color) - Math.min(...color)) / 255;
                return { color, score: bucket.count * (0.5 + saturation) };
            })
            .sort((first, second) => second.score - first.score);

        const selected: number[][] = [];
        for (const candidate of candidates) {
            if (selected.every((color) => {
                const distance = Math.hypot(...color.map((channel, index) => channel - candidate.color[index]));
                return distance >= 65;
            })) selected.push(candidate.color);
            if (selected.length === 6) break;
        }
        return selected.map(([red, green, blue]) => toHex(red, green, blue));
    } finally {
        if (image instanceof ImageBitmap) image.close();
    }
}
