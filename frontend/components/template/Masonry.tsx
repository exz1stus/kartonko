import React, { ReactNode, useEffect, useRef, useState } from "react";
import ec from "clsx";
import MasonryItem from "./MasonryItem";
import { motion, MotionConfig } from "motion/react";

interface Props {
    children: ReactNode
    className?: string
    maxCols?: number
    colWidth?: number
}

interface Column {
    children: ReactNode[]
}

const Masonry: React.FC<Props> = ({ children, className, maxCols = 0, colWidth = 200 }: Props) => {
    const [columns, setColumns] = useState<Column[]>([]);
    const containerRef = useRef<HTMLDivElement>(null);
    const heights = useRef<number[]>([]);
    const hasMeasured = useRef(false);

    const childrenArray = Array.isArray(children) ? children.flat() : [children];


    // const handleMeasure = (height: number) => {
    //     if (heights.current.length >= childrenArray.length) {
    //         return;
    //     }

    //     heights.current.push(height);
    //     if (heights.current.length === childrenArray.length) {
    //         createColumns(childrenArray, columns.length);
    //         hasMeasured.current = true;
    //     }
    // }

    const createColumns = (elements: ReactNode[], count: number) => {
        const cols = Array.from({ length: count }, () => ({ children: [] as ReactNode[] }));
        elements.forEach((el, i) => {
            const colIndex = i % count;
            cols[colIndex].children.push(el);
        });
        return cols;
    };

    useEffect(() => {
        const observer = new ResizeObserver(entries => {
            const width = entries[0].contentRect.width;
            const cols = Math.floor(width / colWidth);
            const clamped = maxCols === 0 ? cols : Math.min(cols, maxCols);
            const finalCount = Math.min(Math.max(1, clamped), childrenArray.length);
            setColumns(createColumns(childrenArray, finalCount));
        });

        if (containerRef.current) observer.observe(containerRef.current);
        return () => observer.disconnect();
    }, [children]);

    const container = true ? (
        columns.map((column, i) => (
            <div key={i} className="flex flex-col gap-4">
                {column.children.map((child) =>
                    child
                )}
            </div>
        ))
    ) : (null
        // columns.map((column, i) => (
        //     <div key={i} className="flex flex-col gap-4">
        //         {childrenArray.map((child, j) => (
        //             <MasonryItem key={j} onMeasure={(h) => handleMeasure(h)}>
        //                 {child}
        //             </MasonryItem>
        //         ))}
        //     </div>
        // ))
    )

    return (
        <div ref={containerRef} className={ec("flex flex-row gap-2", className)} style={{ maxWidth: `${maxCols > 0 && maxCols * colWidth}px` }}>
            {container}
        </div>
    );
};

export default Masonry;