import type { ReactNode } from 'react'
import styles from './Section.module.css'

interface Props {
  title: string
  subtitle?: string
  children: ReactNode
}

export default function Section({ title, subtitle, children }: Props) {
  return (
    <div className={styles.section}>
      <div className={styles.head}>
        <h2 className={styles.title}>{title}</h2>
        {subtitle && <p className={styles.sub}>{subtitle}</p>}
      </div>
      <div className={styles.body}>{children}</div>
    </div>
  )
}
