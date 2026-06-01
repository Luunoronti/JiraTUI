using System;
using System.Security.Cryptography;
using System.Text;

namespace JiraTUI.Config
{
    public static class SecretProtector
    {
        public static string Protect(string plaintext)
        {
            if (string.IsNullOrEmpty(plaintext))
                return "";

            var bytes = Encoding.UTF8.GetBytes(plaintext);
            var protectedBytes = ProtectedData.Protect(bytes, null, DataProtectionScope.CurrentUser);
            return Convert.ToBase64String(protectedBytes);
        }

        public static string Unprotect(string protectedBase64)
        {
            if (string.IsNullOrEmpty(protectedBase64))
                return "";

            try
            {
                var protectedBytes = Convert.FromBase64String(protectedBase64);
                var bytes = ProtectedData.Unprotect(protectedBytes, null, DataProtectionScope.CurrentUser);
                return Encoding.UTF8.GetString(bytes);
            }
            catch
            {
                return "";
            }
        }
    }
}
