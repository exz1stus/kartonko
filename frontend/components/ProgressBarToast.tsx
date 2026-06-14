import React from "react";
import { Progress } from "./ui/progress";

interface Props {
    value: number;
    currentName: string;
}

const ProgressBarToast = ({ value, currentName }: Props) => {
    return (
        <div className="flex flex-col gap-2 pt-1 w-full">
            <div className="flex font-medium text-xs">
                <span className="max-w-45 truncate">{currentName}</span>
            </div>
            <Progress value={value} className="h-1" />
        </div>
    );
};

export default ProgressBarToast;
