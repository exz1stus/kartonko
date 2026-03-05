import { TagEntryData, LogEntryData, ParseLogData } from "@/lib/log";

interface Props {
    data: LogEntryData;
}

const TagDeleted = ({ data }: Props) => {
    const entryData = ParseLogData<TagEntryData>(data.data);
    const description = entryData ? (
        <>{entryData.name}</>
    ) : (
        "error retrieving tag data"
    );
    return <div className="flex justify-around">deleted tag {description}</div>;
};
export default TagDeleted;
