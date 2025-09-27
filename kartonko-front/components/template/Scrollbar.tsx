import React from "react";

interface Props {
    children: React.ReactNode;
    onScroll?: React.UIEventHandler<HTMLDivElement>;
    className?: string;
};

const Scrollbar: React.FC<Props> = ({ children, onScroll, className }) => {
    return (
        <div
            onScroll={onScroll}
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
    )
}

export default Scrollbar
