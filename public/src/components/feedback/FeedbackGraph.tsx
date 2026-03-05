import React from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { AssignmentFeedback } from '../../../proto/qf/types_pb'

interface FeedbackGraphProps {
    feedbacks: AssignmentFeedback[]
    title?: string
}

export const FeedbackGraph: React.FC<FeedbackGraphProps> = ({
    feedbacks,
    title = "Time Spent Distribution",
}) => {
    const getTimeDistribution = () => {
        const timeBuckets: Record<string, { count: number; color: string }> = {}

        // Create buckets for each hour up to 10h+
        for (let i = 0; i < 10; i++) {
            timeBuckets[`${i}-${i + 1}h`] = { count: 0, color: i % 2 === 0 ? "#007bff" : "#6c757d" }
        }
        timeBuckets["10h+"] = { count: 0, color: "#212529" }


        feedbacks.forEach(fb => {
            if (fb.TimeSpent === 0) {
                return
            }
            const hours = fb.TimeSpent / 60
            const hourBucket = Math.floor(hours)
            if (hourBucket < 10) {
                timeBuckets[`${hourBucket}-${hourBucket + 1}h`].count++
            } else {
                timeBuckets["10h+"].count++
            }
        })

        return Object.entries(timeBuckets).map(([range, data]) => ({
            range,
            count: data.count,
            color: data.color
        }))
    }

    const getAverageTime = (): string => {
        const validFeedbacks = feedbacks.filter(fb => fb.TimeSpent > 0)
        if (validFeedbacks.length === 0) return "0h 0m"

        const avgMinutes = validFeedbacks.reduce((sum, fb) => sum + fb.TimeSpent, 0) / validFeedbacks.length
        const hours = Math.floor(avgMinutes / 60)
        const minutes = Math.floor(avgMinutes % 60)

        if (hours > 0 && minutes > 0) return `${hours}h ${minutes}m`
        if (hours > 0) return `${hours}h`
        return `${minutes}m`
    }

    const timeDistribution = getTimeDistribution()
    const totalResponses = feedbacks.filter(fb => fb.TimeSpent > 0).length

    if (timeDistribution.length === 0) {
        return (
            <div className="card bg-base-100 shadow-lg">
                <div className="card-body">
                    <div className="flex items-center gap-3 mb-4">
                        <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
                            <i className="fa fa-chart-bar text-primary" />
                        </div>
                        <h5 className="text-xl font-bold text-base-content">{title}</h5>
                    </div>
                    <div className="flex flex-col items-center justify-center py-12 text-base-content/50">
                        <i className="fa fa-clock-o text-5xl mb-4" />
                        <p className="text-lg">No time data available yet</p>
                    </div>
                </div>
            </div>
        )
    }

    return (
        <div className="card bg-base-100 shadow-lg">
            <div className="card-body">
                <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
                    <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
                            <i className="fa fa-chart-bar text-primary" />
                        </div>
                        <h5 className="text-xl font-bold text-base-content">{title}</h5>
                    </div>
                    <div className="flex flex-wrap gap-4 text-sm text-base-content/70">
                        <div className="flex items-center gap-2 bg-base-200 px-3 py-1 rounded-full">
                            <i className="fa fa-users" />
                            <span className="font-semibold">{totalResponses}</span>
                            <span>responses</span>
                        </div>
                        <div className="flex items-center gap-2 bg-base-200 px-3 py-1 rounded-full">
                            <i className="fa fa-clock-o" />
                            <span>Avg: <span className="font-semibold">{getAverageTime()}</span></span>
                        </div>
                    </div>
                </div>
                <div className="bg-base-200/50 rounded-lg p-4">
                    <ResponsiveContainer width="100%" height={300}>
                        <BarChart data={timeDistribution} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
                            <CartesianGrid strokeDasharray="3 3" />
                            <XAxis
                                dataKey="range"
                                tick={{ fontSize: 11 }}
                                angle={-45}
                                textAnchor="end"
                                height={60}
                            />
                            <YAxis
                                tick={{ fontSize: 12 }}
                                allowDecimals={false}
                            />
                            <Tooltip
                                formatter={(value: number) => [`${value} student${value !== 1 ? "s" : ""}`, "Count"]}
                                labelFormatter={(label: string) => `Time range: ${label}`}
                                contentStyle={{
                                    backgroundColor: "#f8f9fa",
                                    border: "1px solid #dee2e6",
                                    borderRadius: "0.25rem"
                                }}
                            />
                            <Bar
                                dataKey="count"
                                fill="#007bff"
                                radius={[4, 4, 0, 0]}
                            />
                        </BarChart>
                    </ResponsiveContainer>
                </div>
            </div>
        </div>
    )
}

export default FeedbackGraph
