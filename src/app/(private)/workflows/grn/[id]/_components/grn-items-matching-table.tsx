'use client'

interface ReceivedItem {
  id: string
  itemNumber: number
  description: string
  poQuantity: number
  receivedQuantity: number
  unit: string
  variance: number
  damage: number
  damageNotes?: string
  condition: 'GOOD' | 'DAMAGED' | 'PARTIAL'
}

interface GRNItemsMatchingTableProps {
  items: ReceivedItem[]
}

const CONDITION_BADGE: Record<string, string> = {
  GOOD: 'bg-green-100 text-green-800',
  DAMAGED: 'bg-red-100 text-red-800',
  PARTIAL: 'bg-yellow-100 text-yellow-800',
}

export function GRNItemsMatchingTable({ items }: GRNItemsMatchingTableProps) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead className="border-b bg-muted/50">
          <tr>
            <th className="text-left font-semibold py-3 px-4">#</th>
            <th className="text-left font-semibold py-3 px-4">Description</th>
            <th className="text-center font-semibold py-3 px-4">PO Qty</th>
            <th className="text-center font-semibold py-3 px-4">Received</th>
            <th className="text-center font-semibold py-3 px-4">Variance</th>
            <th className="text-center font-semibold py-3 px-4">Damaged</th>
            <th className="text-center font-semibold py-3 px-4">Condition</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id} className="border-b hover:bg-muted/30">
              <td className="py-3 px-4 text-muted-foreground">{item.itemNumber}</td>
              <td className="py-3 px-4">
                <div>
                  <p className="font-medium">{item.description}</p>
                  {item.damageNotes && (
                    <p className="text-xs text-muted-foreground mt-1">{item.damageNotes}</p>
                  )}
                </div>
              </td>
              <td className="py-3 px-4 text-center font-semibold">
                {item.poQuantity} {item.unit}
              </td>
              <td className="py-3 px-4 text-center font-semibold">
                {item.receivedQuantity} {item.unit}
              </td>
              <td className={`py-3 px-4 text-center font-semibold ${
                item.variance !== 0
                  ? item.variance > 0
                    ? 'text-green-600'
                    : 'text-red-600'
                  : ''
              }`}>
                {item.variance > 0 ? '+' : ''}{item.variance}
              </td>
              <td className="py-3 px-4 text-center">
                {item.damage > 0 && (
                  <span className="font-semibold text-red-600">{item.damage}</span>
                )}
                {item.damage === 0 && (
                  <span className="text-muted-foreground">-</span>
                )}
              </td>
              <td className="py-3 px-4 text-center">
                <span className={`px-2 py-1 rounded text-xs font-medium ${CONDITION_BADGE[item.condition]}`}>
                  {item.condition}
                </span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
