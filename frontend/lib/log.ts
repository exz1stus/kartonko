export interface LogEntryData {
    id: number;
    user_id: number;
    entry_type: string;
    affected_obj_id: number;
    created_at: string;
    data: Object | null;
}

export interface ImageEntryData {
    name: string;
}

export interface TagEntryData {
    name: string;
}

export function ParseLogData<T>(data: Object | null): T | null {
    if (!data) return null;
    return data as T;
}
