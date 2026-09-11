"use client";
import { getImageIdId } from "@/lib/api/generated/client";
import { ImageMetadata } from "@/lib/api/generated/model";
import { LogEntryData, ParseLogData, ImageEntryData } from "@/lib/log";
import Link from "next/link";
import { useEffect, useState } from "react";

interface Props {
    data: LogEntryData;
}

const ImageCreated = ({ data }: Props) => {
    const entryData = ParseLogData<ImageEntryData>(data.data);
    const [image, setImage] = useState<ImageMetadata | null>(null);

    const fetchImageData = async () => {
        const img = await getImageIdId(data.affected_obj_id);
        setImage(img);
    };

    useEffect(() => {
        fetchImageData();
    }, [data]);

    const description = entryData ? (
        image ? (
            <Link className="text-blue-400" href={`/image/${image.filename}`}>
                {image.filename}
            </Link>
        ) : (
            <span className="line-through hover:cursor-not-allowed">
                {entryData.name}
            </span>
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
