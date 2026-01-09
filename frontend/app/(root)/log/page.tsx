import LogEntryData from "@/app/lib/log";
import LogEntry from "@/components/Log/LogEntry";

const API_ORIGIN = process.env.NEXT_PUBLIC_API_LOCAL;

const Log = async () => {
    let entriesData: LogEntryData[] = [];

    try {
        const res = await fetch(`${API_ORIGIN}/log?cursor=0&limit=10`);
        entriesData = await res.json();
    } catch (err) {
        console.error("Failed to fetch log:", err);
    }

    const entries = entriesData.map((entry, index) => {
        return <LogEntry key={index} data={entry} />;
    });

    return <div className="flex flex-col gap-2 m-2">{entries}</div>;
};

export default Log;
