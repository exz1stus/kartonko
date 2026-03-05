"use client";
import { LogEntryData, ParseLogData, TagEntryData } from "@/lib/log";
import Link from "next/link";
interface Props {
    data: LogEntryData;
}

const TagCreated = ({ data }: Props) => {
    const entryData = ParseLogData<TagEntryData>(data.data);

    const description = entryData ? (
        <Link className="text-blue-400" href={`/tag/${entryData.name}`}>
            {entryData.name}
        </Link>
    ) : (
        <span className="italic">error retrieving tag data</span>
    );
    return (
        <div className="flex">
            <span>created tag {description}</span>
        </div>
    );
};

export default TagCreated;
