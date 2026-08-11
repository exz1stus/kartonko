export default interface ImageMetadata {
    id: number;
    hash: string;
    filename: string;
    tags: string[];
    format: string;
    width: number;
    height: number;
    user_id: number;
    uploaded_at: string;
}

export interface ImagePostRequest {
    name: string;
    tags: string[];
}

export interface ImagePostBatchRequest {
    data: ImagePostRequest[];
    common_tags: string[];
}

export interface ImageError {
    name: string;
    error: string;
}

export interface ImagePostBatchResponse {
    successes: ImageMetadata[];
    failures: ImageError[];
}
