import { Clock3, MapPinned, NotebookPen, Sparkles } from "lucide-react"
import { cn } from "@/lib/utils"

export function ScenicFieldNote({
  location,
  bestTime,
  mood,
  tag,
  className,
}: Readonly<{
  location: string
  bestTime: string
  mood: string
  tag: string
  className?: string
}>) {
  const copy = {
    title: "Field note",
    locationLabel: "Location",
    windowLabel: "Light window",
    moodLabel: "Mood",
  }

  return (
    <div className={cn("falzo-field-note", className)}>
      <div className="falzo-field-note-header">
        <p className="falzo-field-note-title">
          <NotebookPen className="size-3.5" />
          {copy.title}
        </p>
        <span className="falzo-field-note-pill">{tag}</span>
      </div>

      <div className="falzo-field-note-grid">
        <div className="falzo-field-note-item">
          <p className="falzo-field-note-label inline-flex items-center gap-1">
            <MapPinned className="size-3.5" />
            {copy.locationLabel}
          </p>
          <p className="falzo-field-note-value">{location}</p>
        </div>

        <div className="falzo-field-note-item">
          <p className="falzo-field-note-label inline-flex items-center gap-1">
            <Clock3 className="size-3.5" />
            {copy.windowLabel}
          </p>
          <p className="falzo-field-note-value">{bestTime}</p>
        </div>

        <div className="falzo-field-note-item">
          <p className="falzo-field-note-label inline-flex items-center gap-1">
            <Sparkles className="size-3.5" />
            {copy.moodLabel}
          </p>
          <p className="falzo-field-note-value">{mood}</p>
        </div>
      </div>
    </div>
  )
}
