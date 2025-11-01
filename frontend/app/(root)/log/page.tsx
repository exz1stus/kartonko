
interface LogEntry {
    ID: number;
    user_id: number;
    entryTypeName: string;
    affected_obj_id: number;
    data: string;
}

const API_ORIGIN = process.env.NEXT_PUBLIC_API_LOCAL;

const Log = async () => {
    try {
        const res = await fetch(`${API_ORIGIN}/log?cursor=0&limit=10`);
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
    catch (err) {
        console.error("Failed to fetch log:", err);
        return null;
    }
}

export default Log
