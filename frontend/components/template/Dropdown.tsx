import React from "react";

interface DropdownProps {
    text?: string;
    open: boolean;
    children: React.ReactNode;
}

const Dropdown: React.FC<DropdownProps> = ({
    text,
    open,
    children,
}: DropdownProps) => {
    const [isOpen, setOpen] = React.useState(open);
    return (
        <div className="relative flex flex-col items-center font-bold">
            <span
                className="text-xl hover:cursor-pointer"
                onClick={() => setOpen(!isOpen)}
            >
                {text}
            </span>
            <div
                className="flex flex-col gap-1 px-2 transition-all duration-300 ease-in-out"
                style={{
                    transform: isOpen ? "scaleY(1)" : "scaleY(0)",
                    transformOrigin: "top right",
                    opacity: isOpen ? 1 : 0,
                }}
            >
                {isOpen && children}
            </div>
        </div>
    );
};

export default Dropdown;
