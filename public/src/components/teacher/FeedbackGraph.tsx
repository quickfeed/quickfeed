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
            <div className="card">
                <div className="card-header">
                    <h5 className="mb-0">
                        <i className="fa fa-chart-bar mr-2"></i>
                        {title}
                    </h5>
                </div>
                <div className="card-body text-center text-muted">
                    <i className="fa fa-clock-o fa-3x mb-3"></i>
                    <p>No time data available yet</p>
                </div>
            </div>
        )
    }

    return (
        <div className="card w-100 m-3">
            <div className="card-header d-flex justify-content-between align-items-center">
                <h5 className="mb-0">
                    <i className="fa fa-chart-bar mr-2"></i>
                    {title}
                </h5>
                <div className="text-muted small">
                    <span className="mr-3">
                        <i className="fa fa-users mr-1"></i>
                        {totalResponses} responses
                    </span>
                    <span>
                        <i className="fa fa-clock-o mr-1"></i>
                        Avg: {getAverageTime()}
                    </span>
                </div>
            </div>
            <div className="card-body">
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
    )
}

export default FeedbackGraph
