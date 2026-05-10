interface PublicChatNavbarProps {
  groupName?: string
}

export function PublicChatNavbar({ groupName }: PublicChatNavbarProps) {
  return (
    <div className='border-b-2 border-neutral-950 bg-white p-4 flex items-center justify-between shadow-hard-sm z-10'>
      <h3 className='text-xl font-black text-neutral-950'>
        {groupName}
      </h3>
      <div className='flex items-center gap-2'>
        <span className='text-xs font-bold uppercase tracking-wider text-neutral-600'>Chat Público</span>
        <div className='w-2 h-2 bg-green-500 rounded-full' />
      </div>
    </div>
  )
}
