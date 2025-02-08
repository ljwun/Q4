'use client';

import React from 'react';
import { useEffect, useRef, useState } from 'react';

interface GridCalculatorProps extends React.HTMLAttributes<HTMLDivElement> {
    minItemWidth: string;
    maxItemWidth: string;
    gap: string;
    children?: React.ReactNode;
    className?: string;
    limitRows?: number;
}

// 將不同單位轉換為像素
const convertToPixels = (value: string): number => {
    // 建立一個臨時元素來計算實際像素值
    const temp = document.createElement('div');
    temp.style.visibility = 'hidden';
    temp.style.position = 'absolute';
    document.body.appendChild(temp);

    try {
        const match = value.match(/^([\d.]+)(px|rem|em)$/);
        if (!match) {
            // 預設視為 rem
            temp.style.width = `${parseFloat(value)}rem`;
        } else {
            temp.style.width = value;
        }

        // 獲取計算後的實際像素值
        const pixels = parseFloat(getComputedStyle(temp).width);
        return pixels;
    } finally {
        document.body.removeChild(temp);
    }
};

// 由於現在需要 DOM 環境，我們需要在組件中處理初始值
export const EnhancedGridContainer = ({
    minItemWidth = '30px',
    maxItemWidth = '40px',
    gap = '0px',
    className = '',
    limitRows,
    children,
    ...props
}: GridCalculatorProps) => {
    const containerRef = useRef<HTMLDivElement>(null);
    const [itemWidth, setItemWidth] = useState<number>(0);
    const [minWidthPixels, setMinWidthPixels] = useState<number>(0);
    const [maxWidthPixels, setMaxWidthPixels] = useState<number>(0);
    const [gapPixels, setGapPixels] = useState<number>(0);
    const [childrenArray, setChildrenArray] = useState(React.Children.toArray(children));

    // 初始化像素值
    useEffect(() => {
        setChildrenArray(React.Children.toArray(children));
        setMinWidthPixels(convertToPixels(minItemWidth));
        setMaxWidthPixels(convertToPixels(maxItemWidth));
        setGapPixels(convertToPixels(gap));
        setItemWidth(convertToPixels(maxItemWidth));
    }, [minItemWidth, maxItemWidth, gap, children]);

    useEffect(() => {
        if (!containerRef.current) return;

        const observer = new ResizeObserver((entries) => {
            for (const entry of entries) {
                // 如果沒有設定最大最小寬度，則不進行計算
                if (!maxWidthPixels || !minWidthPixels) {
                    console.warn('Max width or min width not set, skipping calculation.');
                    return;
                };
                // 取得容器寬度
                const containerWidth = entry.contentRect.width;
                // 計算可能的列數範圍（使用最大寬度計算）
                let columns: number;
                if (maxWidthPixels) {
                    columns = Math.ceil((containerWidth + gapPixels) / (maxWidthPixels + gapPixels));
                } else {
                    columns = Math.floor((containerWidth + gapPixels) / (minWidthPixels + gapPixels));
                }
                // 依照列數反推實際的元素寬度
                const calculatedWidth = (containerWidth - (gapPixels * (columns - 1))) / columns;
                // 檢查是否在範圍內
                if (minWidthPixels && calculatedWidth < minWidthPixels || maxWidthPixels && calculatedWidth > maxWidthPixels) {
                    console.warn(`Calculated width: ${calculatedWidth}px not in range [${minWidthPixels}px, ${maxWidthPixels}px]`);
                }
                setItemWidth(calculatedWidth);
                // 如果有限制行數，則只顯示指定行數的元素
                if (limitRows) {
                    setChildrenArray(React.Children.toArray(children).slice(0, limitRows * columns));
                };
            }
        });

        observer.observe(containerRef.current);

        return () => {
            observer.disconnect();
        };
    }, [minItemWidth, maxItemWidth, gap, maxWidthPixels, minWidthPixels, gapPixels, limitRows, children, childrenArray]);

    return (
        <div
            ref={containerRef}
            // 如果有purgecss，這裡的class名稱需要在tailwind.config.js中設定safelist
            className={`grid gap-[var(--custom-gap)] grid-cols-1 md:grid-cols-[var(--custom-item-width)] ${className}`}
            style={{
                '--custom-gap': gap,
                '--custom-item-width': `repeat(auto-fit,${itemWidth}px)`
            } as React.CSSProperties}
            {...props}
        >
            <>{childrenArray}</>
        </div>
    );
};