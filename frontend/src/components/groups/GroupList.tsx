import type { Group } from "../../types";
import { GroupCard } from "./GroupCard";

interface GroupListProps {
    groups: Group[]
    onToggle: (id: number, is_active: boolean) => void
    onDelete: (id: number) => void
    onCopy?: (slug: string) => void
    loadingIds?: number[]
}

export function GroupList({ groups, onToggle, onDelete, onCopy, loadingIds = [] }: GroupListProps) {
    return (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {groups.map(group => (
                <GroupCard key={group.id} group={group} onToggle={onToggle} onDelete={onDelete} onCopy={onCopy} isLoading={loadingIds.includes(group.id)} />
            ))}
        </div>
    )
}