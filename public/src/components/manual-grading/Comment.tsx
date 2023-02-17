import React from "react"
import ReactMarkdown from "react-markdown"
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { CodeProps } from "react-markdown/lib/ast-to-react";


const CriterionComment = ({comment}: {comment: string}): JSX.Element | null => {
    if (comment == "" || comment.length == 0) {
        return null
    }
    return (
        <div className="comment-md">
        <ReactMarkdown
            // eslint-disable-next-line react/no-children-prop
            children={comment}
            components={{
                // eslint-disable-next-line @typescript-eslint/no-unused-vars
                code({node, inline, className, children, ...props}: CodeProps) {
                  const match = /language-(\w+)/.exec(className || '')
                  return !inline && match ? (
                    <SyntaxHighlighter
                        // eslint-disable-next-line react/no-children-prop
                        children={String(children).replace(/\n$/, '')}
                        language={match[1]}
                        PreTag="div"
                        {...props}
                        style={oneDark}
                    />
                  ) : (
                    <code className={className} {...props}>
                      {children}
                    </code>
                  )
                }
              }}
            />
        </div>
    )
}

export default CriterionComment