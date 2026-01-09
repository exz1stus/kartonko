import React from "react";
import ec from "clsx";

interface Props {
    word: string;
    className?: string;
}

const FancySpan = ({ word, className }: Props) => {
    return (
        <p className={ec(className)}>
            {word.split("").map((letter, index) => (
                <span
                    key={index}
                    className="inline-block hover:scale-200 transition-transform duration-200"
                >
                    {letter}
                </span>
            ))}
        </p>
    );
};

export default FancySpan;
