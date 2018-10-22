# undo_log

## undo_log 格式
开始 transaction：
{START, tID, 0, 0}

记录 transaction 涉及的 user 的原始信息：
{UPDATE, tID, userID, cash}

开始 undolog 的 checkpoint：
{STARTCHECKPOINT, 0, 0, 0}

完成 undolog 的 checkpoint：
{ENDCHECKPOINT, 0, 0, 0}

## 回滚 transaction
从后向前遍历系统的 undolog，直到遇到当前 transaction 的开始标志时停止遍历，在过程中将涉及用户的 cash 值修改为undolog中记录的 cash值。

## 回收历史 undolog 日志
系统每 500 ms 对 undolog 日志进行一次回收，回收规则：从后向前遍历系统的 undolog， 如果遍历到 checkpoint 完成标志，则将接下来遍历到的第一个 checkpoint 开始标志之前的 undolog 回收。

回收完成后系统开始对 undolog 进行新一轮的 checkpoint, checkpoint 规则：系统向 undolog 内写入 checkpoint 开始标志并记录当前系统中活跃的 transaction，系统等待这些 transaction 全部完成然后向 undolog 内写入 checkpoint 完成标志。

## 锁
系统获取锁的顺序： 系统全局锁（如果需要） -> userid 较小的用户锁 -> userid 较大的用户锁