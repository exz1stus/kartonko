import { getUserServer } from '@/app/lib/auth'

interface LogEntry {
    ID: number;
    user_id: number;
    entryTypeName: string;
    affected_obj_id: number;
    data: string;
}

const API_LOCAL = process.env.NEXT_PUBLIC_API_LOCAL;

const Log = async () => {
    const user = await getUserServer();

    const res = await fetch(`${API_LOCAL}/auditlog?cursor=0&limit=10`, { method: "GET" });
    const log: { entries: LogEntry[] } = await res.json();
    const entries = log.entries.map((entry: LogEntry, index: number) => (
        <div key={index}>
            <p>{entry?.ID || "Unknown"}</p>
            <p>{entry?.data || "No data"}</p>
        </div>
    ));
    return (
        <div>
            {entries}
        </div>
    )
}

export default Log
