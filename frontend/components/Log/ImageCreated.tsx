import LogEntryData from "@/app/lib/log";
import React from "react";

interface Props {
    data: LogEntryData;
}

const ImageCreated = ({ data }: Props) => {
    return (
        <div className="flex justify-around">
            <div>{data.affected_obj_id}</div>
            <div>{data.user_id}</div>
            <div>{data.created_at}</div>
        </div>
    );
};

export default ImageCreated;
