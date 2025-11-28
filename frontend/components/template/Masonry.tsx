import React, { ReactNode, use, useEffect, useMemo, useRef, useState } from "react";
import ec from "clsx";
import { useDebouncedCallback } from "use-debounce";

interface Props {
    children: ReactNode
    ratios: number[]
    className?: string
    maxCols?: number
    colWidth?: number
}

interface Column {
    children: ReactNode[]
    height: number
}

const Masonry: React.FC<Props> = ({ children, className, ratios, maxCols = 0, colWidth = 200 }: Props) => {
    const [columns, setColumns] = useState<Column[]>([]);
    const containerRef = useRef<HTMLDivElement>(null);

    const childrenArray = useMemo(
        () => (Array.isArray(children) ? children.flat() : [children]),
        [children]
    );

    const getColumnWidth = (count: number) => {
        if (maxCols > 0) return colWidth;
        return (containerRef.current?.clientWidth || 0) / count;
    };

    const createColumns = (elements: ReactNode[], count: number) => {
        const cols = Array.from({ length: count }, () => ({ children: [] as ReactNode[], height: 0 }));
        elements.forEach((el, i) => {
            const h = ratios[i] * getColumnWidth(count);
            const targetCol = cols.reduce(
                (min, c) => (c.height < min.height ? c : min),
                cols[0]
            );
            targetCol.children.push(el);
            targetCol.height += h;
        });
        return cols;
    }

    useEffect(() => {
        const observer = new ResizeObserver(entries => {
            const width = entries[0].contentRect.width;
            const cols = Math.floor(width / colWidth);
            const clamped = maxCols === 0 ? cols : Math.min(cols, maxCols);
            const finalCount = Math.min(Math.max(1, clamped), childrenArray.length);
            setColumns(createColumns(childrenArray, finalCount));
        });

        if (containerRef.current) observer.observe(containerRef.current);
        return () => observer.disconnect()

    }, [childrenArray, colWidth, maxCols, ratios]);

    return (
        <div ref={containerRef} className={ec("flex flex-row gap-2", className)} style={{ maxWidth: maxCols > 0 ? `${maxCols * colWidth}px` : "100%" }}>
            {columns.map((column, i) => (
                <div key={i} className="flex flex-col gap-4"
                    style={{ maxWidth: (containerRef.current?.clientWidth || colWidth) / columns.length }}>
                    {column.children.map((child) =>
                        child
                    )}
                </div>
            ))
            }
        </div >
    );
};

export default Masonry;