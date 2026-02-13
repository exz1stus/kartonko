import Link from "next/link";
import React from "react";

const Logo = () => {
    return (
        <Link
            href="/"
            className="left-1/2 absolute font-bold text-2xl -translate-x-1/2 cursor-pointer transform"
        >
            <span>kartonko</span>
        </Link>
    );
};

export default Logo;
