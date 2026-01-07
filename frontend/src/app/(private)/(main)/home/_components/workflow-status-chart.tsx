"use client";

import * as React from "react";
import { Label, Pie, PieChart, Sector } from "recharts";
import { type PieSectorDataItem } from "recharts/types/polar/Pie";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ChartContainer,
  ChartStyle,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DashboardMetrics } from "@/types";

interface WorkflowStatusChartProps {
  metrics: DashboardMetrics;
}

const chartConfig = {
  documents: {
    label: "Documents",
  },
  draft: {
    label: "Draft",
    color: "var(--chart-1)",
  },
  submitted: {
    label: "Submitted",
    color: "var(--chart-2)",
  },
  inApproval: {
    label: "In Approval",
    color: "var(--chart-3)",
  },
  approved: {
    label: "Approved",
    color: "var(--chart-5)", // Will be defined as green in CSS
  },
  rejected: {
    label: "Rejected",
    color: "var(--chart-5)",
  },
} satisfies ChartConfig;

export function WorkflowStatusChart({ metrics }: WorkflowStatusChartProps) {
  const id = "workflow-status-chart";

  // Transform metrics data into chart format
  const workflowData = React.useMemo(() => [
    {
      status: "draft",
      documents: metrics.draftDocuments || 0,
      fill: "var(--color-draft)",
    },
    {
      status: "submitted",
      documents: metrics.submittedDocuments || 0,
      fill: "var(--color-submitted)",
    },
    {
      status: "inApproval",
      documents: metrics.pendingApproval || 0,
      fill: "var(--color-inApproval)",
    },
    {
      status: "approved",
      documents: metrics.approvedDocuments || 0,
      fill: "var(--color-approved)",
    },
    {
      status: "rejected",
      documents: metrics.rejectedDocuments || 0,
      fill: "var(--color-rejected)",
    },
  ].filter((item) => item.documents > 0), [metrics]);

  const [activeStatus, setActiveStatus] = React.useState(
    workflowData.length > 0 ? workflowData[0].status : "draft"
  );

  const statuses = React.useMemo(
    () => workflowData.map((item) => item.status),
    [workflowData]
  );

  if (workflowData.length === 0) {
    return (
      <Card data-chart={id} className="flex flex-col">
        <CardHeader>
          <CardTitle className="text-base font-bold">Workflow Status</CardTitle>
          <CardDescription>No documents found</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-1 items-center justify-center">
          <p className="text-muted-foreground">No workflow data available</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card data-chart={id} className="flex flex-col">
      <ChartStyle id={id} config={chartConfig} />
      <CardHeader className="flex-row items-start space-y-0 pb-0">
        <div className="grid gap-">
          <CardTitle className="text-base font-bold">Workflow Tasks</CardTitle>
          <CardDescription className="text-xs font-normal">Document distribution by status</CardDescription>
        </div>
        <Select value={activeStatus} onValueChange={setActiveStatus}>
          <SelectTrigger
            className="ml-auto h-7 w-[130px] rounded-lg pl-2.5"
            aria-label="Select a status"
          >
            <SelectValue placeholder="Select status" />
          </SelectTrigger>
          <SelectContent align="end" className="rounded-xl">
            {statuses.map((key) => {
              const config = chartConfig[key as keyof typeof chartConfig];
              if (!config) {
                return null;
              }
              return (
                <SelectItem
                  key={key}
                  value={key}
                  className="rounded-lg [&_span]:flex"
                >
                  <div className="flex items-center gap-2 text-xs">
                    <span
                      className="flex h-3 w-3 shrink-0 rounded-xs"
                      style={{
                        backgroundColor: `var(--color-${key})`,
                      }}
                    />
                    {config?.label}
                  </div>
                </SelectItem>
              );
            })}
          </SelectContent>
        </Select>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col justify-center pb-6">
        <ChartContainer
          id={id}
          config={chartConfig}
          className="mx-auto aspect-square w-full max-w-[300px]"
        >
          <PieChart>
            <ChartTooltip
              cursor={false}
              content={<ChartTooltipContent hideLabel />}
            />
            <Pie
              data={workflowData}
              dataKey="documents"
              nameKey="status"
              innerRadius={60}
              strokeWidth={5}
              activeShape={({
                outerRadius = 0,
                ...props
              }: PieSectorDataItem) => (
                <g>
                  <Sector {...props} outerRadius={outerRadius + 10} />
                  <Sector
                    {...props}
                    outerRadius={outerRadius + 25}
                    innerRadius={outerRadius + 12}
                  />
                </g>
              )}
            >
              <Label
                content={({ viewBox }) => {
                  if (viewBox && "cx" in viewBox && "cy" in viewBox) {
                    const activeData = workflowData.find(item => item.status === activeStatus);
                    const displayValue = activeData?.documents || 0;
                    const displayLabel = chartConfig[activeStatus as keyof typeof chartConfig]?.label || "Documents";
                    
                    return (
                      <text
                        x={viewBox.cx}
                        y={viewBox.cy}
                        textAnchor="middle"
                        dominantBaseline="middle"
                      >
                        <tspan
                          x={viewBox.cx}
                          y={viewBox.cy}
                          className="fill-foreground text-3xl font-bold"
                        >
                          {displayValue.toLocaleString()}
                        </tspan>
                        <tspan
                          x={viewBox.cx}
                          y={(viewBox.cy || 0) + 24}
                          className="fill-muted-foreground"
                        >
                          {displayLabel}
                        </tspan>
                      </text>
                    );
                  }
                }}
              />
            </Pie>
          </PieChart>
        </ChartContainer>
        
        {/* Legend */}
        <div className="mt-4 mb-2 grid grid-cols-2 gap-2 text-sm">
          {workflowData.map((item) => {
            const config = chartConfig[item.status as keyof typeof chartConfig];
            return (
              <div
                key={item.status}
                className={`flex items-center gap-2 rounded-lg p-2 transition-colors ${
                  activeStatus === item.status 
                    ? 'bg-muted/50 border border-border' 
                    : 'hover:bg-muted/30'
                }`}
                onClick={() => setActiveStatus(item.status)}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    setActiveStatus(item.status);
                  }
                }}
              >
                <span
                  className="flex h-3 w-3 shrink-0 rounded-full"
                  style={{
                    backgroundColor: `var(--color-${item.status})`,
                  }}
                />
                <span className="flex-1 font-medium">
                  {config?.label}
                </span>
                <span className="font-bold text-foreground">
                  {item.documents}
                </span>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
