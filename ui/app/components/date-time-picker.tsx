"use client";

import * as React from "react";
import { CalendarIcon } from "@radix-ui/react-icons";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";

interface DateTimePickerProps extends React.InputHTMLAttributes<HTMLInputElement> {
  defaultDate?: Date
  label?: string
}

export function DateTimePicker({ defaultDate, label = "MM/DD/YYYY hh:mm aa", ...props }: DateTimePickerProps) {
  const [date, setDate] = React.useState<Date | undefined>(defaultDate);
  const [isOpen, setIsOpen] = React.useState(false);
  const hourScrollRef = React.useRef<HTMLDivElement>(null);
  const minuteScrollRef = React.useRef<HTMLDivElement>(null);

  const hours = Array.from({ length: 12 }, (_, i) => i === 0 ? 12 : i);
  const minutes = Array.from({ length: 60 }, (_, i) => i);

  // 初始滾動位置設定
  React.useEffect(() => {
    if (date && isOpen) {
      setTimeout(() => {
        if (hourScrollRef.current) {
          const hourElement = hourScrollRef.current.querySelector(`[data-hour="${date.getHours() % 12 || 12}"]`);
          hourElement?.scrollIntoView({ block: 'center', behavior: 'auto' });
        }
        if (minuteScrollRef.current) {
          const minuteElement = minuteScrollRef.current.querySelector(`[data-minute="${date.getMinutes()}"]`);
          minuteElement?.scrollIntoView({ block: 'center', behavior: 'auto' });
        }
      }, 0);
    }
  }, [date, isOpen]);

  // 處理滾動重置
  const handleScroll = (ref: React.RefObject<HTMLDivElement | null>, itemHeight: number, totalItems: number) => {
    if (!ref.current) return;
    
    const container = ref.current;
    const scrollTop = container.scrollTop;
    const viewportHeight = container.clientHeight;
    const contentHeight = itemHeight * totalItems;
    const threshold = itemHeight * 2; // 設定觸發重置的閾值

    // 當滾動到頂部附近時
    if (scrollTop < threshold) {
      container.scrollTop = scrollTop + contentHeight / 3;
    }
    // 當滾動到底部附近時
    else if (scrollTop > contentHeight / 3 * 2 - viewportHeight) {
      container.scrollTop = scrollTop - contentHeight / 3;
    }
  };

  React.useEffect(() => {
    const hourContainer = hourScrollRef.current;
    const minuteContainer = minuteScrollRef.current;

    if (hourContainer && minuteContainer) {
      const handleHourScroll = () => handleScroll(hourScrollRef, 40, 36); // 12小時 * 3組
      const handleMinuteScroll = () => handleScroll(minuteScrollRef, 40, 180); // 60分鐘 * 3組

      hourContainer.addEventListener('scroll', handleHourScroll);
      minuteContainer.addEventListener('scroll', handleMinuteScroll);

      return () => {
        hourContainer.removeEventListener('scroll', handleHourScroll);
        minuteContainer.removeEventListener('scroll', handleMinuteScroll);
      };
    }
  }, []);

  const handleDateSelect = (selectedDate: Date | undefined) => {
    if (selectedDate) {
      setDate(selectedDate);
    }
  };

  const handleTimeChange = (
    type: "hour" | "minute" | "ampm",
    value: string,
    section?: 'top' | 'middle' | 'bottom'
  ) => {
    if (!date) return;
    
    const newDate = new Date(date);
    if (type === "hour") {
      const hour = parseInt(value);
      newDate.setHours(
        (hour % 12) + (newDate.getHours() >= 12 ? 12 : 0)
      );
    } else if (type === "minute") {
      newDate.setMinutes(parseInt(value));
    } else if (type === "ampm") {
      const currentHours = newDate.getHours();
      newDate.setHours(
        value === "PM" ? 
          (currentHours % 12) + 12 : 
          currentHours % 12
      );
    }
    setDate(newDate);

    // 根據點擊區域調整滾動位置
    if (section) {
      const ref = type === "hour" ? hourScrollRef : minuteScrollRef;
      const container = ref.current;
      if (container) {
        const itemHeight = 40; // 按鈕高度
        const currentScroll = container.scrollTop;
        let adjustment = 0;
        
        if (section === 'top') {
          adjustment = itemHeight * 12; // 向下滾動一組
        } else if (section === 'bottom') {
          adjustment = -itemHeight * 12; // 向上滾動一組
        }
        
        container.scrollTop = currentScroll + adjustment;
      }
    }
  };

  return (
    <Popover open={isOpen} onOpenChange={(open) => {
      if (!date && open) setDate(new Date())
      setIsOpen(open)
    }}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className={cn(
            "w-full justify-start text-left font-normal",
            !date && "text-muted-foreground"
          )}
        >
          <CalendarIcon className="mr-2 h-4 w-4" />
          {date ? (
            date.toLocaleString()
          ) : (
            <span>{label}</span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0">
        <div className="sm:flex">
          <Calendar
            mode="single"
            selected={date}
            onSelect={handleDateSelect}
            initialFocus
          />
          <div className="flex flex-col sm:flex-row sm:h-[300px] divide-y sm:divide-y-0 sm:divide-x">
            <ScrollArea className="w-64 sm:w-auto">
              <div ref={hourScrollRef} className="flex sm:flex-col p-2 sm:pb-[150px] sm:pt-[150px]">
                {/* 頂部小時 */}
                {hours.map((hour) => (
                  <Button
                    key={`top-${hour}`}
                    size="icon"
                    variant={
                      date && (date.getHours() % 12 || 12) === hour
                        ? "default"
                        : "ghost"
                    }
                    className="sm:w-full shrink-0 aspect-square"
                    onClick={() => handleTimeChange("hour", hour.toString(), 'top')}
                  >
                    {hour}
                  </Button>
                ))}
                {/* 中間小時 */}
                {hours.map((hour) => (
                  <Button
                    key={`middle-${hour}`}
                    data-hour={hour}
                    size="icon"
                    variant={
                      date && (date.getHours() % 12 || 12) === hour
                        ? "default"
                        : "ghost"
                    }
                    className="sm:w-full shrink-0 aspect-square"
                    onClick={() => handleTimeChange("hour", hour.toString(), 'middle')}
                  >
                    {hour}
                  </Button>
                ))}
                {/* 底部小時 */}
                {hours.map((hour) => (
                  <Button
                    key={`bottom-${hour}`}
                    size="icon"
                    variant={
                      date && (date.getHours() % 12 || 12) === hour
                        ? "default"
                        : "ghost"
                    }
                    className="sm:w-full shrink-0 aspect-square"
                    onClick={() => handleTimeChange("hour", hour.toString(), 'bottom')}
                  >
                    {hour}
                  </Button>
                ))}
              </div>
              <ScrollBar orientation="horizontal" className="sm:hidden" />
            </ScrollArea>
            <ScrollArea className="w-64 sm:w-auto">
              <div ref={minuteScrollRef} className="flex sm:flex-col p-2 sm:pb-[150px] sm:pt-[150px]">
                {/* 頂部分鐘 */}
                {minutes.map((minute) => (
                  <Button
                    key={`top-${minute}`}
                    size="icon"
                    variant={
                      date && date.getMinutes() === minute
                        ? "default"
                        : "ghost"
                    }
                    className="sm:w-full shrink-0 aspect-square"
                    onClick={() => handleTimeChange("minute", minute.toString(), 'top')}
                  >
                    {minute}
                  </Button>
                ))}
                {/* 中間分鐘 */}
                {minutes.map((minute) => (
                  <Button
                    key={`middle-${minute}`}
                    data-minute={minute}
                    size="icon"
                    variant={
                      date && date.getMinutes() === minute
                        ? "default"
                        : "ghost"
                    }
                    className="sm:w-full shrink-0 aspect-square"
                    onClick={() => handleTimeChange("minute", minute.toString(), 'middle')}
                  >
                    {minute}
                  </Button>
                ))}
                {/* 底部分鐘 */}
                {minutes.map((minute) => (
                  <Button
                    key={`bottom-${minute}`}
                    size="icon"
                    variant={
                      date && date.getMinutes() === minute
                        ? "default"
                        : "ghost"
                    }
                    className="sm:w-full shrink-0 aspect-square"
                    onClick={() => handleTimeChange("minute", minute.toString(), 'bottom')}
                  >
                    {minute}
                  </Button>
                ))}
              </div>
              <ScrollBar orientation="horizontal" className="sm:hidden" />
            </ScrollArea>
            <ScrollArea className="">
              <div className="flex sm:flex-col p-2">
                {["AM", "PM"].map((ampm) => (
                  <Button
                    key={ampm}
                    size="icon"
                    variant={
                      date &&
                        ((ampm === "AM" && date.getHours() < 12) ||
                          (ampm === "PM" && date.getHours() >= 12))
                        ? "default"
                        : "ghost"
                    }
                    className="sm:w-full shrink-0 aspect-square"
                    onClick={() => handleTimeChange("ampm", ampm)}
                  >
                    {ampm}
                  </Button>
                ))}
              </div>
            </ScrollArea>
          </div>
        </div>
      </PopoverContent>
      <input
        type="hidden"
        {...props}
        value={date?.toISOString() || ''}
      />
    </Popover>
  );
}