import React from 'react'

interface FeedbackSortControlsProps {
    sortOrder: 'asc' | 'desc' | 'none'
    setSortOrder: (order: 'asc' | 'desc' | 'none') => void
    feedbackCount: number
}

export const FeedbackSortControls: React.FC<FeedbackSortControlsProps> = ({
    sortOrder,
    setSortOrder,
    feedbackCount
}) => {
    const toggleSort = () => {
        if (sortOrder === 'none') setSortOrder('asc')
        else if (sortOrder === 'asc') setSortOrder('desc')
        else setSortOrder('none')
    }

    const getSortIcon = () => {
        if (sortOrder === 'asc') return 'fa-sort-amount-asc'
        if (sortOrder === 'desc') return 'fa-sort-amount-desc'
        return 'fa-sort'
    }

    const getSortLabel = () => {
        if (sortOrder === 'asc') return 'Time (Low to High)'
        if (sortOrder === 'desc') return 'Time (High to Low)'
        return 'Sort by Time'
    }

    return (
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6 p-4 bg-base-200 rounded-lg">
            <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
                    <i className="fa fa-list text-primary" />
                </div>
                <div>
                    <h3 className="text-xl font-bold text-base-content">Individual Feedback</h3>
                    <span className="text-sm text-base-content/60">{feedbackCount} responses received</span>
                </div>
            </div>
            <button
                className={`btn btn-sm gap-2 ${sortOrder !== 'none' ? 'btn-primary' : 'btn-outline'}`}
                onClick={toggleSort}
            >
                <i className={`fa ${getSortIcon()}`} />
                {getSortLabel()}
            </button>
        </div>
    )
}

export default FeedbackSortControls
