import React from 'react'

interface FancySpanProps {
    word: string
}

const FancySpan = ({ word }: FancySpanProps) => {
    return (
        <p>
            {word.split("").map((letter, index) => (
                <span
                    key={index}
                    className="inline-block transition-transform duration-200 hover:scale-200"
                >
                    {letter}
                </span>
            ))}
        </p>
    )
}

export default FancySpan
