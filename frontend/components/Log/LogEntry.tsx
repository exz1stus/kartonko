import React from "react";
import ImageCreated from "./ImageCreated";
import LogEntryData from "@/app/lib/log";

interface Props {
    data: LogEntryData;
}

const LogEntry = ({ data }: Props) => {
    const entryComponents: Record<string, (data: LogEntryData) => React.JSX.Element> = {
        image_created: (data) => <ImageCreated data={data} />,
    };

    const entry = entryComponents[data.entry_type] ? (
        entryComponents[data.entry_type](data)
    ) : (
        <div>Unknown entry type: {data.entry_type}</div>
    );

    return <div className="bg-surface-0 border-1 rounded-l">{entry}</div>;
};

export default LogEntry;
