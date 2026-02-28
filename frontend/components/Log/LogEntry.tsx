import ImageCreated from "./ImageCreated";
import {LogEntryData} from "@/lib/log";
import UserElement from "../UserElement";
import TimeField from "../TimeField";
import ImageDeleted from "./ImageDeleted";

interface Props {
    data: LogEntryData;
}

const LogEntry = ({ data }: Props) => {
    const entryComponents: Record<
        string,
        (data: LogEntryData) => React.JSX.Element
    > = {
        image_created: (data) => <ImageCreated data={data} />,
        image_deleted: (data) => <ImageDeleted data={data} />,
    };

    const entry = entryComponents[data.entry_type] ? (
        entryComponents[data.entry_type](data)
    ) : (
        <div>Unknown entry type: {data.entry_type}</div>
    );

    return (
        <div className="flex justify-between items-center bg-surface-0 p-2 border rounded-l">
            <div className="flex flex-row items-center gap-2">
                <UserElement id={data.user_id} />
                {entry}
            </div>
            <TimeField time={data.created_at} />
        </div>
    );
};

export default LogEntry;
