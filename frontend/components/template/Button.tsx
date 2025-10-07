import React from 'react'

interface ButtonProps {
    text: string;
    onClick: () => void;
}

const Button: React.FC<ButtonProps> = ({ text, onClick }: ButtonProps) => {
    return (
        <button
            className="px-3 rounded-2xl border-1 border-surface-30 w-full h-full cursor-pointer"
            onClick={onClick}
        >
            {text}
        </button>
    )
}

export default Button
