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
    minCols?: number;
    colWidthPx?: number;
    gap?: number;
}

interface Column {
    items: MasonryItem[];
    height: number;
}

const Masonry: React.FC<Props> = ({
    items,
    className,
    maxCols = 0,
    minCols = 1,
    colWidthPx = 200,
    gap = 16,
}: Props) => {
    const [columns, setColumns] = useState<Column[]>([]);
    const containerRef = useRef<HTMLDivElement>(null);

    const childrenArray = useMemo(
        () => (Array.isArray(items) ? items.flat() : [items]),
        [items],
    );

    const createColumns = (
        items: MasonryItem[],
        count: number,
        columnWidth: number,
    ) => {
        const cols = Array.from({ length: count }, () => ({
            items: [] as MasonryItem[],
            height: 0,
        }));
        items.forEach((item) => {
            const h = item.ratio * columnWidth;
            const targetCol = cols.reduce(
                (min, c) => (c.height < min.height ? c : min),
                cols[0],
            );
            targetCol.items.push(item);
            targetCol.height += h + 1;
        });
        return cols;
    };

    useEffect(() => {
        if (!containerRef.current) return;

        const observer = new ResizeObserver((entries) => {
            const width = entries[0].contentRect.width;
            if (!width || childrenArray.length === 0) {
                setColumns([]);
                return;
            }

            // colWidthPx is the preferred minimum width. Include the gap when
            // deciding how many columns fit in the available content width.
            const preferredCols = Math.floor(
                (width + gap) / (colWidthPx + gap),
            );
            const cappedCols =
                maxCols > 0 ? Math.min(preferredCols, maxCols) : preferredCols;
            const finalCount = Math.min(
                childrenArray.length,
                Math.max(1, minCols, cappedCols),
            );
            const columnWidth = (width - gap * (finalCount - 1)) / finalCount;

            setColumns(createColumns(childrenArray, finalCount, columnWidth));
        });

        if (containerRef.current) observer.observe(containerRef.current);
        return () => observer.disconnect();
    }, [childrenArray, colWidthPx, gap, maxCols, minCols]);

    return (
        <div
            ref={containerRef}
            className={ec("grid w-full h-full", className)}
            style={{
                columnGap: `${gap}px`,
                gridTemplateColumns: `repeat(${columns.length}, minmax(0, 1fr))`,
            }}
        >
            {columns.map((column, i) => (
                <div key={i} className="flex flex-col gap-5 min-w-0">
                    {column.items.map((item) => (
                        <React.Fragment key={item.key}>
                            {item.item}
                        </React.Fragment>
                    ))}
                </div>
            ))}
        </div>
    );
};

export default Masonry;
export type { MasonryItem };
