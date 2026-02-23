import React, { ReactNode, useEffect, useMemo, useRef, useState } from "react";
import ec from "clsx";

interface MasonryItem {
    key: string;
    item: ReactNode;
    ratio: number;
}

interface Props {
    items: MasonryItem[];
    className?: string;
    maxCols?: number;
    colWidth?: number;
}

interface Column {
    items: MasonryItem[];
    height: number;
}

const Masonry: React.FC<Props> = ({ items, className, maxCols = 0, colWidth = 200 }: Props) => {
    const [columns, setColumns] = useState<Column[]>([]);
    const containerRef = useRef<HTMLDivElement>(null);

    const childrenArray = useMemo(() => (Array.isArray(items) ? items.flat() : [items]), [items]);

    const createColumns = (items: MasonryItem[], count: number, width: number) => {
        const cols = Array.from({ length: count }, () => ({
            items: [] as MasonryItem[],
            height: 0,
        }));
        items.forEach((item) => {
            const h = item.ratio * width;
            const targetCol = cols.reduce((min, c) => (c.height < min.height ? c : min), cols[0]);
            targetCol.items.push(item);
            targetCol.height += h + 1;
        });
        return cols;
    };

    useEffect(() => {
        if (!containerRef.current) return;

        const observer = new ResizeObserver((entries) => {
            const width = entries[0].contentRect.width;
            if (!width) return;
            const cols = Math.floor(width / colWidth);
            const clamped = maxCols === 0 ? cols : Math.min(cols, maxCols);
            const finalCount = Math.min(Math.max(1, clamped), childrenArray.length);
            setColumns(createColumns(childrenArray, finalCount, width));
        });

        if (containerRef.current) observer.observe(containerRef.current);
        return () => observer.disconnect();
    }, [childrenArray, colWidth, maxCols]);

    return (
        <div
            ref={containerRef}
            className={ec("flex-row justify-center flex gap-4 w-full", className)}
            style={{ maxWidth: maxCols > 0 ? `${maxCols * colWidth}px` : "100%" }}
        >
            {columns.map((column, i) => (
                <div
                    key={i}
                    className={`flex flex-col gap-5 lg:max-w-[25vw]`}
                    style={{
                        width: `${100 / columns.length}%`,
                    }}
                >
                    {column.items.map((item) => (
                        <React.Fragment key={item.key}>{item.item}</React.Fragment>
                    ))}
                </div>
            ))}
        </div>
    );
};

export default Masonry;
export type { MasonryItem };
