import React from "react";

interface Props {
    children: React.ReactNode;
    className?: string;
}

const Scrollbar: React.FC<Props> = ({ children, className }) => {
    return (
        <div
            className={`
                h-full overflow-y-auto
                scrollbar-thin
                scrollbar-thumb-primary-0
                scrollbar-hover:scrollbar-thumb-primary-0
                active:scrollbar-thumb-primary-0
                scrollbar-track-transparent
                ${className ?? ""}
            `}
        >
            {children}
        </div>
    );
};

export default Scrollbar;
