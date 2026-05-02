"use client";
import * as React from "react";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Line,
  LineChart,
  Area,
  AreaChart,
  XAxis,
  YAxis,
  Cell,
} from "recharts";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  ChartLegend,
  ChartLegendContent,
  type ChartConfig,
} from "@/components/ui/chart";
import { cn } from "@/lib/utils";

export type ReportChartKind = "bar" | "line" | "area";

export interface ReportSeries<T = Record<string, unknown>> {
  /** Key in each data row to plot. Constrained to `keyof T` when `T` is supplied. */
  dataKey: keyof T & string;
  /** Display label. */
  label: string;
  /** Optional explicit color override. Defaults to chart-1..5 cycling. */
  color?: string;
}

export interface ReportChartProps<T extends Record<string, unknown> = Record<string, unknown>> {
  kind: ReportChartKind;
  data: T[];
  /** Key in each data row to use as the X axis label. */
  xKey: keyof T & string;
  series: ReportSeries<T>[];
  /** Tailwind classes for the outer container; defaults to a sensible aspect. */
  className?: string;
  /** Show legend. Default false (single-series charts don't need it). */
  showLegend?: boolean;
  /** For bar charts: per-bar color from palette instead of single hue. Default false. */
  perBarColor?: boolean;
}

const PALETTE = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
];

export function ReportChart<T extends Record<string, unknown>>({
  kind,
  data,
  xKey,
  series,
  className,
  showLegend,
  perBarColor,
}: ReportChartProps<T>) {
  const config: ChartConfig = React.useMemo(() => {
    const cfg: ChartConfig = {};
    series.forEach((s, i) => {
      cfg[s.dataKey] = {
        label: s.label,
        color: s.color || PALETTE[i % PALETTE.length],
      };
    });
    return cfg;
  }, [series]);

  return (
    <ChartContainer config={config} className={cn("aspect-[16/7] w-full", className)}>
      {kind === "bar" ? (
        <BarChart data={data}>
          <CartesianGrid strokeDasharray="3 3" vertical={false} />
          <XAxis dataKey={xKey} tickLine={false} axisLine={false} />
          <YAxis tickLine={false} axisLine={false} width={32} />
          <ChartTooltip content={<ChartTooltipContent />} />
          {showLegend && <ChartLegend content={<ChartLegendContent />} />}
          {series.map((s) => (
            <Bar
              key={s.dataKey}
              dataKey={s.dataKey}
              fill={`var(--color-${s.dataKey})`}
              radius={[4, 4, 0, 0]}
            >
              {perBarColor &&
                data.map((_, idx) => (
                  <Cell key={idx} fill={PALETTE[idx % PALETTE.length]} />
                ))}
            </Bar>
          ))}
        </BarChart>
      ) : kind === "line" ? (
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" vertical={false} />
          <XAxis dataKey={xKey} tickLine={false} axisLine={false} />
          <YAxis tickLine={false} axisLine={false} width={32} />
          <ChartTooltip content={<ChartTooltipContent />} />
          {showLegend && <ChartLegend content={<ChartLegendContent />} />}
          {series.map((s) => (
            <Line
              key={s.dataKey}
              type="monotone"
              dataKey={s.dataKey}
              stroke={`var(--color-${s.dataKey})`}
              strokeWidth={2}
              dot={false}
            />
          ))}
        </LineChart>
      ) : (
        <AreaChart data={data}>
          <defs>
            {series.map((s) => (
              <linearGradient
                key={s.dataKey}
                id={`fill-${s.dataKey}`}
                x1="0"
                y1="0"
                x2="0"
                y2="1"
              >
                <stop
                  offset="5%"
                  stopColor={`var(--color-${s.dataKey})`}
                  stopOpacity={0.4}
                />
                <stop
                  offset="95%"
                  stopColor={`var(--color-${s.dataKey})`}
                  stopOpacity={0}
                />
              </linearGradient>
            ))}
          </defs>
          <CartesianGrid strokeDasharray="3 3" vertical={false} />
          <XAxis dataKey={xKey} tickLine={false} axisLine={false} />
          <YAxis tickLine={false} axisLine={false} width={32} />
          <ChartTooltip content={<ChartTooltipContent />} />
          {showLegend && <ChartLegend content={<ChartLegendContent />} />}
          {series.map((s) => (
            <Area
              key={s.dataKey}
              type="monotone"
              dataKey={s.dataKey}
              stroke={`var(--color-${s.dataKey})`}
              strokeWidth={2}
              fill={`url(#fill-${s.dataKey})`}
            />
          ))}
        </AreaChart>
      )}
    </ChartContainer>
  );
}
