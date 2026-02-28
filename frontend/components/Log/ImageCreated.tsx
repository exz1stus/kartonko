import { getImageMetadataByIdServer } from "@/lib/image";
import { LogEntryData, ParseLogData, ImageEntryData } from "@/lib/log";
import Link from "next/link";

interface Props {
    data: LogEntryData;
}

const ImageCreated = async ({ data }: Props) => {
    const entryData = ParseLogData<ImageEntryData>(data.data);
    const image = await getImageMetadataByIdServer(data.affected_obj_id);

    const description = entryData ? (
        image ? (
            <Link className="text-blue-400" href={`/image/${image.filename}`}>{image.filename}</Link>
        ) : (
            <span className="line-through hover:cursor-not-allowed"> {entryData.name} </span>
        )
    ) : (
        <span className="italic">error retrieving image data</span>
    );
    return (
        <div className="flex">
            <span>uploaded image {description}</span>
        </div>
    );
};

export default ImageCreated;
