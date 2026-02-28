import {LogEntryData} from "@/lib/log";
import LogEntry from "@/components/Log/LogEntry";
import { serverFetch } from "@/lib/serverFetch";
import AuthGuard from "@/components/AuthGuard";
import Scrollbar from "@/components/template/Scrollbar";

const Log = async () => {
    let entriesData: LogEntryData[] = [];

    try {
        const res = await serverFetch(`/log?cursor=0&limit=100`);
        entriesData = await res.json();
    } catch (err) {
        console.error("Failed to fetch log:", err);
    }

    const entries = entriesData.map((entry, index) => {
        return (
            <div key={index} className="w-full lg:max-w-[50%]">
                <LogEntry data={entry} />
            </div>
        );
    });

    return (
        <AuthGuard moderator={true}>
            <Scrollbar>
                <div className="flex flex-col items-center gap-2 m-2">{entries}</div>
            </Scrollbar>
        </AuthGuard>
    );
};

export default Log;
