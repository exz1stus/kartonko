import { LucideLoader } from "lucide-react";
import React from "react";

const Loading = () => {
    return (
        <div className="flex justify-center items-center w-full h-full">
            <LucideLoader />
        </div>
    );
};

export default Loading;
