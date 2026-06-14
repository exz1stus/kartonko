const MAXIMUM_UPLOAD_SIZE = 50 * 1024 * 1024; //50MB

export function isUploadSizeValid(size: number): boolean {
    return size <= MAXIMUM_UPLOAD_SIZE;
}
