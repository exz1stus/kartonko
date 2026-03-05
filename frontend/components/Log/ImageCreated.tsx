"use client";
import ImageMetadata from "@/lib/image";
import { getImageMetadataById } from "@/lib/image.client";
import { LogEntryData, ParseLogData, ImageEntryData } from "@/lib/log";
import Link from "next/link";
import { useEffect, useEffectEvent, useState } from "react";

interface Props {
    data: LogEntryData;
}

const ImageCreated = ({ data }: Props) => {
    const entryData = ParseLogData<ImageEntryData>(data.data);
    const [image, setImage] = useState<ImageMetadata | null>(null);

    const fetchImageData = async () => {
        const img = await getImageMetadataById(data.affected_obj_id);
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
