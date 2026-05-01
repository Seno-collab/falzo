# `src` Developer Guide

## Muc tieu
- Giam thoi gian onboard cho dev moi.
- Giu cau truc de doc, de tim, de fix.
- Dung mot kieu router duy nhat: Next.js App Router.

## Cau truc thu muc
- `app/`: App Router + app-level providers/layout.
- `features/`: logic theo domain/chuc nang. Moi feature tu quan ly `api`, `types`, `screens`, va data rieng neu co.
- `components/`: reusable UI (`ui`, `layout`, `feedback`, `scenic`).
- `lib/`: utility va route constants.
- `i18n/`: message dictionary.
- `types/`: shared type dung chung giua nhieu feature.
- `generated/`: file sinh tu script.

## Quy uoc quan trong
1. Khong hardcode URL route trong screen/component.
2. Dung route constants tai `lib/routes.ts`.
3. Route file trong `app/**/page.tsx` chi nen la route entry mong, import screen tu `features/*/screens`.
4. Component co dung hook/browser API phai co `"use client"`.
5. Bien moi truong uu tien `NEXT_PUBLIC_*` (co the giu fallback cu khi can migrate).
6. Khong giu route group/file rong hoac component khong duoc su dung.

## Pattern tao route moi
1. Them route key trong `lib/routes.ts`.
2. Tao screen trong `features/<feature>/screens/<name>-screen.tsx`.
3. Tao `app/<segment>/page.tsx` va render screen do.
4. Neu co UI dung lai giua nhieu feature, tach sang `components/`.
5. Neu can link/redirect, import `ROUTES` thay vi ghi chuoi.

## Checklist maintain
1. Moi route phai co owner ro rang trong `features/<feature>`.
2. Khong tao 2 bo component cung muc dich (tranh duplicate layer).
3. Context provider chi them khi it nhat 1 route dang dung.
4. Truoc merge, chay `pnpm typecheck` va `pnpm build`.
