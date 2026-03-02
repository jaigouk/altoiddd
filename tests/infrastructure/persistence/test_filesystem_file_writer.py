"""Tests for FilesystemFileWriter adapter."""

from __future__ import annotations

import os
import stat
from typing import TYPE_CHECKING

import pytest

if TYPE_CHECKING:
    from pathlib import Path

from src.application.ports.file_writer_port import FileWriterPort
from src.infrastructure.persistence.filesystem_file_writer import FilesystemFileWriter


class TestFilesystemFileWriter:
    """Tests for FilesystemFileWriter."""

    def test_write_file_creates_file(self, tmp_path: Path) -> None:
        writer = FilesystemFileWriter()
        target = tmp_path / "output.md"
        writer.write_file(target, "hello world")
        assert target.read_text(encoding="utf-8") == "hello world"

    def test_write_file_creates_parent_dirs(self, tmp_path: Path) -> None:
        writer = FilesystemFileWriter()
        target = tmp_path / "a" / "b" / "c" / "file.md"
        writer.write_file(target, "nested content")
        assert target.read_text(encoding="utf-8") == "nested content"

    def test_write_file_overwrites_existing(self, tmp_path: Path) -> None:
        writer = FilesystemFileWriter()
        target = tmp_path / "file.md"
        target.write_text("old content", encoding="utf-8")
        writer.write_file(target, "new content")
        assert target.read_text(encoding="utf-8") == "new content"

    def test_write_file_empty_content(self, tmp_path: Path) -> None:
        writer = FilesystemFileWriter()
        target = tmp_path / "empty.md"
        writer.write_file(target, "")
        assert target.read_text(encoding="utf-8") == ""

    def test_write_file_unicode_content(self, tmp_path: Path) -> None:
        writer = FilesystemFileWriter()
        target = tmp_path / "unicode.md"
        content = "日本語テスト -- U o a n"
        writer.write_file(target, content)
        assert target.read_text(encoding="utf-8") == content

    def test_write_file_permission_error(self, tmp_path: Path) -> None:
        writer = FilesystemFileWriter()
        readonly_dir = tmp_path / "readonly"
        readonly_dir.mkdir()
        os.chmod(readonly_dir, stat.S_IRUSR | stat.S_IXUSR)
        target = readonly_dir / "file.md"
        try:
            with pytest.raises(PermissionError):
                writer.write_file(target, "content")
        finally:
            os.chmod(readonly_dir, stat.S_IRWXU)

    def test_satisfies_file_writer_port_protocol(self) -> None:
        writer = FilesystemFileWriter()
        assert isinstance(writer, FileWriterPort)
