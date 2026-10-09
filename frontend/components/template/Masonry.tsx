import React, { ReactNode, useEffect, useMemo, useRef, useState } from "react";
import ec from "clsx";

interface MasonryItem {
    key: string;
    item: ReactNode;
    ratio: number;
    extraHeightPx?: number;
}

interface Props {
    items: MasonryItem[];
    className?: string;
    maxCols?: number;
    minCols?: number;
    colWidthPx?: number;
    maxColWidthPx?: number;
    gap?: number;
    rowGapPx?: number;
}

interface Column {
    items: MasonryItem[];
    height: number;
}

const createColumns = (
    items: MasonryItem[],
    count: number,
    columnWidth: number,
    rowGapPx: number,
): Column[] => {
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
        targetCol.height += h + (item.extraHeightPx ?? 0) + rowGapPx;
    });
    return cols;
};

const Masonry: React.FC<Props> = ({
    items,
    className,
    maxCols = 0,
    minCols = 1,
    colWidthPx = 200,
    maxColWidthPx = 0,
    gap = 16,
    rowGapPx = 20,
}: Props) => {
    const [columns, setColumns] = useState<Column[]>([]);
    const containerRef = useRef<HTMLDivElement>(null);

    const childrenArray = useMemo(
        () => (Array.isArray(items) ? items.flat() : [items]),
        [items],
    );

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

            setColumns(
                createColumns(childrenArray, finalCount, columnWidth, rowGapPx),
            );
        });

        if (containerRef.current) observer.observe(containerRef.current);
        return () => observer.disconnect();
    }, [childrenArray, colWidthPx, gap, maxCols, minCols, rowGapPx]);

    const limitedColumnCount =
        maxCols > 0
            ? Math.min(childrenArray.length, maxCols)
            : childrenArray.length;
    const maxWidthStyle =
        maxColWidthPx > 0 && limitedColumnCount > 0
            ? `${limitedColumnCount * maxColWidthPx + Math.max(0, limitedColumnCount - 1) * gap}px`
            : maxCols > 0
              ? `${(maxCols + 1) * colWidthPx}px`
              : undefined;

    return (
        <div
            ref={containerRef}
            className={ec("grid w-full h-full", className)}
            style={{
                maxWidth: maxWidthStyle,
                columnGap: `${gap}px`,
                gridTemplateColumns: `repeat(${columns.length}, minmax(0, 1fr))`,
            }}
        >
            {columns.map((column, i) => (
                <div
                    key={i}
                    className="flex flex-col min-w-0"
                    style={{ rowGap: `${rowGapPx}px` }}
                >
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
