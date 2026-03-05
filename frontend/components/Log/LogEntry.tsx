"use client";
import { LogEntryData } from "@/lib/log";
import UserElement from "../UserElement";
import TimeField from "../TimeField";
import { UserData } from "@/lib/user";
import { getUserById } from "@/lib/user.client";
import { useCallback, useEffect, useEffectEvent, useState } from "react";

import ImageCreated from "./ImageCreated";
import ImageDeleted from "./ImageDeleted";
import TagCreated from "./TagCreated";
import TagDeleted from "./TagDeleted";

interface Props {
    data: LogEntryData;
}

const LogEntry = ({ data }: Props) => {
    const [user, setUser] = useState<UserData | null>(null);
    const entryComponents: Record<
        string,
        (data: LogEntryData) => React.JSX.Element
    > = {
        image_created: (data) => <ImageCreated data={data} />,
        image_deleted: (data) => <ImageDeleted data={data} />,
        tag_created: (data) => <TagCreated data={data} />,
        tag_deleted: (data) => <TagDeleted data={data} />,
    };

    const entry = entryComponents[data.entry_type] ? (
        entryComponents[data.entry_type](data)
    ) : (
        <div>Unknown entry type: {data.entry_type}</div>
    );

    const fetchUser = async () => {
        const user = await getUserById(data.user_id);
        setUser(user);
    };

    useEffect(() => {
        fetchUser();
    }, [data]);

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
