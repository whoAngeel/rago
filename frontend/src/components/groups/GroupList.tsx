import type { Group } from "../../types";
import { GroupCard } from "./GroupCard";

interface GroupListProps {
    groups: Group[]
    onToggle: (id: number, is_active: boolean, name: string) => void
    onDelete: (id: number) => void
    loadingIds?: number[]
}

export function GroupList({ groups, onToggle, onDelete, loadingIds = [] }: GroupListProps) {
    return (
        <div className="grid grid-cols-2 gap-4">
            {groups.map(group => (
                <GroupCard key={group.id} group={group} onToggle={onToggle} onDelete={onDelete} isLoading={loadingIds.includes(group.id)} />
            ))}
        </div>
    )
}