interface CardProps {
    children: React.ReactNode
    padding?: string
    className?: string
}

export const Card = ({
    children,
    padding = "p-4",
    className = ""
}: CardProps) => {
    return <>
        <div className={`
            ${padding} bg-neutral-0 border border-neutral-900 
            rounded-card shadow-hard-lg ${className}`
        } >
            {children}
        </div>
    </>
}