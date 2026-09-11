"use client";
import { LogEntryData } from "@/lib/log";
import UserElement from "../UserElement";
import TimeField from "../TimeField";
import { UserDataResponse } from "@/lib/api/generated/model";
import { useEffect, useState } from "react";

import ImageCreated from "./ImageCreated";
import ImageDeleted from "./ImageDeleted";
import TagCreated from "./TagCreated";
import TagDeleted from "./TagDeleted";
import { getUserIdId } from "@/lib/api/generated/client";

interface Props {
    data: LogEntryData;
}

const LogEntry = ({ data }: Props) => {
    const [user, setUser] = useState<UserDataResponse | null>(null);
    const entryComponents: Record<
        string,
        (data: LogEntryData) => React.JSX.Element
    > = {
        image_create: (data) => <ImageCreated data={data} />,
        image_delete: (data) => <ImageDeleted data={data} />,
        tag_create: (data) => <TagCreated data={data} />,
        tag_delete: (data) => <TagDeleted data={data} />,
    };

    const entryType = data.object_type + "_" + data.action;

    const entry = entryComponents[entryType] ? (
        entryComponents[entryType](data)
    ) : (
        <div>Unknown entry type: {entryType}</div>
    );

    const fetchUser = async () => {
        const user = await getUserIdId(data.user_id);
        setUser(user);
    };

    useEffect(() => {
        fetchUser();
    }, [data, fetchUser]);

    return (
        <div className="flex justify-between items-center bg-surface-0 p-2 border hover:border-surface-50 rounded-lg">
            <div className="flex flex-row items-center gap-2">
                <UserElement user={user} />
                {entry}
            </div>
            <TimeField time={data.created_at} />
        </div>
    );
};

export default LogEntry;
