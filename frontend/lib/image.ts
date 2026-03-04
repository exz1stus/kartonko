import { createHash } from "crypto";
import { createReadStream } from "fs";

export default interface ImageMetadata {
    filename: string;
    tags: string[];
    width: number;
    height: number;
    format: string;
    user_id: number;
    uploaded_at: string;
}

export async function hashFile(file: File): Promise<string> {
    const arrayBuffer = await file.arrayBuffer();
    const hashBuffer = await crypto.subtle.digest("SHA-256", arrayBuffer);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray
        .map((b) => b.toString(16).padStart(2, "0"))
        .join("");

    return hashHex;
}
