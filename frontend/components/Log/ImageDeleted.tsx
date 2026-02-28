import { ImageEntryData, LogEntryData, ParseLogData } from "@/lib/log";

interface Props {
    data: LogEntryData;
}

const ImageDeleted = ({ data }: Props) => {
    const entryData = ParseLogData<ImageEntryData>(data.data);
    const description = entryData ? (
        <>
            {entryData.name}
        </>
    ) : (
        "error retrieving image data"
    );
    return (
        <div className="flex justify-around">
            deleted image {description}
        </div>
    );
};
export default ImageDeleted;
