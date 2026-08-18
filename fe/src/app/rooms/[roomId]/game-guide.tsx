"use client";

import { useEffect, useRef, useState } from "react";
import styles from "./room.module.css";

type GameGuideProps = {
  mrWhiteEnabled: boolean;
};

export function GameGuide({ mrWhiteEnabled }: GameGuideProps) {
  const [open, setOpen] = useState(false);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const dialogRef = useRef<HTMLElement>(null);

  useEffect(() => {
    if (!open) return;

    function closeOnEscape(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setOpen(false);
        triggerRef.current?.focus();
      }
    }

    function closeOnOutsideClick(event: PointerEvent) {
      const target = event.target as Node;
      if (!dialogRef.current?.contains(target) && !triggerRef.current?.contains(target)) {
        setOpen(false);
      }
    }

    window.addEventListener("keydown", closeOnEscape);
    window.addEventListener("pointerdown", closeOnOutsideClick);
    return () => {
      window.removeEventListener("keydown", closeOnEscape);
      window.removeEventListener("pointerdown", closeOnOutsideClick);
    };
  }, [open]);

  function closeGuide() {
    setOpen(false);
    window.requestAnimationFrame(() => triggerRef.current?.focus());
  }

  return (
    <>
      <button
        aria-expanded={open}
        aria-haspopup="dialog"
        className={styles.gameGuideButton}
        onClick={() => setOpen((current) => !current)}
        ref={triggerRef}
        type="button"
      >
        <span aria-hidden="true">?</span>
        Cách chơi
      </button>

      {open && (
        <div className={styles.gameGuideOverlay}>
          <section
            aria-labelledby="game-guide-title"
            className={styles.gameGuideDialog}
            ref={dialogRef}
            role="dialog"
          >
            <header className={styles.gameGuideHeader}>
              <div>
                <small>HƯỚNG DẪN NHANH</small>
                <h2 id="game-guide-title">Chơi Falzo như thế nào?</h2>
                <p>Quan sát lời mô tả, tìm người có từ khác và đừng để lộ từ bí mật của bạn.</p>
              </div>
              <button autoFocus aria-label="Đóng hướng dẫn" onClick={closeGuide} type="button">×</button>
            </header>

            <div className={styles.gameGuideBody}>
              <section aria-labelledby="game-flow-title" className={styles.guideSection}>
                <div className={styles.guideSectionTitle}>
                  <span aria-hidden="true">01</span>
                  <div>
                    <small>MỘT VÒNG CHƠI</small>
                    <h3 id="game-flow-title">4 bước để tìm ra người nằm vùng</h3>
                  </div>
                </div>

                <ol className={styles.guideSteps}>
                  <li>
                    <span>1</span>
                    <div><strong>Xem thẻ bí mật</strong><p>Nhớ từ và vai trò của bạn, không cho người khác xem.</p></div>
                  </li>
                  <li>
                    <span>2</span>
                    <div><strong>Lần lượt mô tả</strong><p>Mỗi người đưa một gợi ý liên quan nhưng không nói thẳng từ khóa.</p></div>
                  </li>
                  <li>
                    <span>3</span>
                    <div><strong>Bỏ phiếu</strong><p>Chọn người đáng nghi nhất. Người có nhiều phiếu nhất sẽ bị loại.</p></div>
                  </li>
                  <li>
                    <span>4</span>
                    <div><strong>Kiểm tra kết quả</strong><p>Nếu chưa có phe thắng, những người còn lại tiếp tục mô tả ở vòng mới.</p></div>
                  </li>
                </ol>
              </section>

              <section aria-labelledby="roles-title" className={styles.guideSection}>
                <div className={styles.guideSectionTitle}>
                  <span aria-hidden="true">02</span>
                  <div>
                    <small>VAI TRÒ & CHỨC NĂNG</small>
                    <h3 id="roles-title">Bạn cần làm gì?</h3>
                  </div>
                </div>

                <div className={styles.roleGuideGrid}>
                  <article className={`${styles.roleGuideCard} ${styles.civilianGuide}`}>
                    <div><span aria-hidden="true">◎</span><strong>Dân thường</strong></div>
                    <p><b>Nhìn thấy:</b> cùng một từ khóa với đa số người chơi.</p>
                    <p><b>Nhiệm vụ:</b> mô tả đủ rõ để nhận ra đồng đội, nhưng không làm lộ từ.</p>
                    <small>THẮNG KHI UNDERCOVER VÀ MR. WHITE ĐỀU BỊ LOẠI</small>
                  </article>

                  <article className={`${styles.roleGuideCard} ${styles.undercoverGuide}`}>
                    <div><span aria-hidden="true">◈</span><strong>Undercover</strong></div>
                    <p><b>Nhìn thấy:</b> một từ gần nghĩa nhưng khác từ của Dân thường.</p>
                    <p><b>Nhiệm vụ:</b> hòa nhập, suy luận từ của số đông và tránh bị bỏ phiếu.</p>
                    <small>THẮNG KHI SỐ UNDERCOVER CÒN LẠI BẰNG HOẶC NHIỀU HƠN DÂN THƯỜNG</small>
                  </article>

                  {mrWhiteEnabled && (
                    <article className={`${styles.roleGuideCard} ${styles.mrWhiteGuide}`}>
                      <div><span aria-hidden="true">?</span><strong>Mr. White</strong></div>
                      <p><b>Nhìn thấy:</b> không có từ khóa — bạn phải đoán từ lời mô tả.</p>
                      <p><b>Nhiệm vụ:</b> ứng biến để không bị phát hiện. Khi bị loại, bạn có một lần đoán từ của Dân thường.</p>
                      <small>THẮNG NGAY NẾU ĐOÁN ĐÚNG TỪ KHÓA</small>
                    </article>
                  )}
                </div>
              </section>

              <aside className={styles.guideTips}>
                <span aria-hidden="true">!</span>
                <div>
                  <strong>Gợi ý để ván chơi hấp dẫn hơn</strong>
                  <p>Dùng liên tưởng ngắn và vừa đủ mơ hồ. Không nói, đánh vần, dịch trực tiếp từ khóa hoặc lặp nguyên gợi ý của người trước.</p>
                </div>
              </aside>
            </div>

            <footer className={styles.gameGuideFooter}>
              <span>Cuộn để xem thêm · Nhấn Esc để đóng</span>
              <button onClick={closeGuide} type="button">Đã hiểu, vào chơi</button>
            </footer>
          </section>
        </div>
      )}
    </>
  );
}
