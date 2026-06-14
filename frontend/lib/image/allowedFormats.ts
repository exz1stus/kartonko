const allowedFormats = ["gif", "png", "jpg", "jpeg"];

export default function isAllowed(format: string) {
    return allowedFormats.includes(format);
}
