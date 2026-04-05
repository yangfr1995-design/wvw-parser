import styles from './StatCard.module.css'

interface Props {
  label: string
  value: string | number
  variant?: 'default' | 'danger' | 'success'
}

export default function StatCard({ label, value, variant = 'default' }: Props) {
  return (
    <div className={styles.card}>
      <div className={`${styles.value} ${styles[variant]}`}>{value}</div>
      <div className={styles.label}>{label}</div>
    </div>
  )
}
