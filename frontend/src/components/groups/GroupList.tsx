import type { Group } from "../../types";
import { GroupCard } from "./GroupCard";

interface GroupListProps {
    groups: Group[]
    onToggle: (id: number) => void
    onDelete: (id: number) => void
}

export function GroupList({ groups, onToggle, onDelete }: GroupListProps) {

    console.log(groups.length)
    return (
        <div className="grid grid-cols-2 gap-4">
            {groups.map(group => (
                <GroupCard key={group.id} group={group} onToggle={onToggle} onDelete={onDelete} />
            ))}
        </div>
    )
}