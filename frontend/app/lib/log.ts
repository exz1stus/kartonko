import { Timestamp } from "next/dist/server/lib/cache-handlers/types";

interface LogEntryData {
    id: number;
    user_id: number;
    entry_type: string;
    affected_obj_id: number;
    created_at: Timestamp;
    data: string;
}

export default LogEntryData;
